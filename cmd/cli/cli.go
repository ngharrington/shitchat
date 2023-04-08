package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/grpc"

	pb "github.com/ngharrington/shitchat/message"
)

func main() {
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
