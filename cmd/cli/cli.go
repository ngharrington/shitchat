package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/jroimartin/gocui"
	pb "github.com/ngharrington/shitchat/message"
)

func run() {
	var history []string
	var recent string

	reader := bufio.NewReader(os.Stdin)

	// Add hostname and port flags
	hostname := flag.String("host", "localhost", "hostname of the server")
	port := flag.Int("port", 50051, "port number of the server")
	flag.Parse()

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *hostname, *port), grpc.WithInsecure())
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	client := pb.NewMessageServiceClient(conn)

	stream, err := client.Broadcast(context.Background())
	if err != nil {
		fmt.Println("Error receiving messages from server:", err)
		return
	}

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				fmt.Println("Error receiving message from server:", err)
				return
			}
			history = append(history, msg.Text)
		}
	}()

	for {
		fmt.Printf("\033[1A\033[K> %s\n", recent)

		fmt.Println("Message history:")
		for _, msg := range history {
			fmt.Println("|", msg)
		}

		fmt.Println("\n -------------------")
		fmt.Print("> ")

		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		history = append(history, strings.TrimSpace(text))
		recent = history[len(history)-1]

		if err := stream.Send(&pb.SendMessageRequest{Text: recent}); err != nil {
			fmt.Println("Error sending message to server:", err)
			continue
		}
	}
}

var counter int

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("message", gocui.KeyEnter, gocui.ModNone, handleMessage(g)); err != nil {
		log.Panicln(err)
	}

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

func handleMessage(g *gocui.Gui) func(*gocui.Gui, *gocui.View) error {
	return func(_ *gocui.Gui, v *gocui.View) error {
		message := v.Buffer()
		v.Clear()
		v.SetCursor(0, 0)
		if message != "" {
			historyView, _ := g.View("history")
			fmt.Fprintln(historyView, message)
		}
		return nil
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
