package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"google.golang.org/grpc"

	"github.com/google/uuid"
	"github.com/jroimartin/gocui"
	pb "github.com/ngharrington/shitchat/message"
)

func createClient(host string, port uint64) (pb.MessageServiceClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewMessageServiceClient(conn)
	return client, nil
}

var privateKeyPath string

func init() {
	// Register the --keyfile flag
	flag.StringVar(&privateKeyPath, "keyfile", "", "Path to the private key file")
}

func main() {

	flag.Parse()

	// Validate the required --keyfile flag
	if privateKeyPath == "" {
		log.Fatal("Missing required --keyfile flag")
	}

	client, err := createClient("localhost", 50051)
	if err != nil {
		log.Panic(err)
	}
	stream, err := client.Broadcast(context.Background())
	if err != nil {
		fmt.Println("Error receiving messages from server:", err)
		return
	}
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("message", gocui.KeyEnter, gocui.ModNone, handleMessage(g, stream)); err != nil {
		log.Panicln(err)
	}

	go listenForMessages(g, stream)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("history", 1, 1, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprint(v, "CHAT HISTORY\n\n")
	}
	if v, err := g.SetView("message", 1, maxY-4, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView("message"); err != nil {
			return err
		}
	}
	return nil
}

func handleMessage(g *gocui.Gui, stream pb.MessageService_BroadcastClient) func(*gocui.Gui, *gocui.View) error {
	privateKeyPath := "/home/neal/workspace/shitchat/scratch/keys/key.pem"

	return func(_ *gocui.Gui, v *gocui.View) error {
		id := uuid.New().String()
		message := strings.TrimSpace(v.Buffer())
		v.Clear()
		v.SetCursor(0, 0)
		if message != "" {
			privateKey, err := readPkFromFile(privateKeyPath)
			if err != nil {
				log.Fatalf("Error reading private key: %s\n", err)
				return err
			}

			// Compute the hash of the message
			hashed := sha256.Sum256([]byte(message))

			// Sign the hash
			signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
			if err != nil {
				log.Fatalf("Error from signing: %s\n", err)
				return err
			}

			// Convert the signature to a string so it can be sent
			signatureStr := base64.StdEncoding.EncodeToString(signature)

			// Send the message along with its signature
			stream.Send(&pb.SendMessageRequest{Id: id, Text: message, Signature: signatureStr, Username: "key.pem"})
		}
		return nil
	}
}

func listenForMessages(g *gocui.Gui, stream pb.MessageService_BroadcastClient) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("Error receiving message from server:", err)
			return
		}
		g.Update(func(g *gocui.Gui) error {
			historyView, _ := g.View("history")
			fmt.Fprintln(historyView, msg.Text)
			// Scroll the history view
			_, maxY := historyView.Size()
			linesInBuffer := len(historyView.BufferLines())
			if linesInBuffer > maxY {
				_, err := historyView.Line(linesInBuffer - maxY)
				if err == nil {
					historyView.SetOrigin(0, linesInBuffer-maxY)
				}
			}
			return nil
		})
	}
}

func readPkFromFile(filepath string) (*rsa.PrivateKey, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, errors.New("no valid PEM data found")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rssPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid RSA public key")
	}

	return rssPrivateKey, nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
