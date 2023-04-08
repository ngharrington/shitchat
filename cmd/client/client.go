package main

import (
	"context"
	"log"
	"time"

	"github.com/ngharrington/shitchat/message"

	"google.golang.org/grpc"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	c := message.NewMessageServiceClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.SendMessage(ctx, &message.SendMessageRequest{
		Text: "Hello, gRPC!",
	})
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
	log.Println("Server response:", res.GetStatus())
}
