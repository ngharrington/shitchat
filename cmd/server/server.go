package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/ngharrington/shitchat/message"

	"google.golang.org/grpc"
)

type server struct {
	message.UnimplementedMessageServiceServer
}

func (s *server) SendMessage(ctx context.Context, req *message.SendMessageRequest) (*message.SendMessageResponse, error) {
	go func() {
		fmt.Println("Received message:", req.GetText())
	}()
	return &message.SendMessageResponse{Status: "accepted"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	message.RegisterMessageServiceServer(s, &server{})

	log.Println("Server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
