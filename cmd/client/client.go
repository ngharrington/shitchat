package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/ngharrington/shitchat/message"
	"google.golang.org/grpc"
)

func main() {
	hostname := flag.String("hostname", "localhost", "The server hostname")
	port := flag.String("port", "50051", "The server port")
	flag.Parse()

	address := fmt.Sprintf("%s:%s", *hostname, *port)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	client := message.NewMessageServiceClient(conn)

	stream, err := client.Broadcast(context.Background())
	if err != nil {
		log.Fatalf("Error opening stream: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			res, err := stream.Recv()
			if err != nil {
				log.Fatalf("Error receiving message: %v", err)
			}
			fmt.Println("Received:", res.GetText())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, _ := inputMessage(reader)
			if err := stream.Send(&message.SendMessageRequest{Text: msg}); err != nil {
				log.Fatalf("Error sending message: %v", err)
			}
		}
	}()

	wg.Wait()
}

func inputMessage(reader *bufio.Reader) (string, error) {
	fmt.Print("Enter message: ")
	msg, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(msg), nil
}
