package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/ngharrington/shitchat/message"
	"google.golang.org/grpc"
)

type server struct {
	message.UnimplementedMessageServiceServer

	mu       sync.Mutex
	clients  map[string]message.MessageService_BroadcastServer
	clientID int
}

func (s *server) Broadcast(stream message.MessageService_BroadcastServer) error {
	s.mu.Lock()
	s.clientID++
	clientID := fmt.Sprintf("Client %d", s.clientID)
	s.clients[clientID] = stream
	s.mu.Unlock()

	for {
		msg, err := stream.Recv()
		fmt.Println(msg)
		if err != nil {
			s.mu.Lock()
			delete(s.clients, clientID)
			s.mu.Unlock()
			return err
		}

		s.mu.Lock()
		for id, client := range s.clients {
			if id == clientID {
				continue
			}
			client.Send(&message.SendMessageResponse{Text: fmt.Sprintf("%s: %s", clientID, msg.GetText())})
		}
		s.mu.Unlock()
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	message.RegisterMessageServiceServer(s, &server{clients: make(map[string]message.MessageService_BroadcastServer)})

	log.Println("Server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
