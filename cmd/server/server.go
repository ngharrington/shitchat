package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/ngharrington/shitchat/internal"
	"github.com/ngharrington/shitchat/message"
	"google.golang.org/grpc"
)

type server struct {
	message.UnimplementedMessageServiceServer

	clients       map[string]message.MessageService_BroadcastServer
	clientID      int
	mu            sync.Mutex
	authenticator internal.Authenticator
}

func (s *server) Broadcast(stream message.MessageService_BroadcastServer) error {
	s.mu.Lock()
	s.clientID++
	clientID := fmt.Sprintf("Client %d", s.clientID)
	s.clients[clientID] = stream
	s.mu.Unlock()

	for {
		// TODO: this error handling seems like it is meant to handle the initial connection
		// not sure maybe better handling on the other loops?
		msg, err := stream.Recv()
		if err != nil {
			s.mu.Lock()
			delete(s.clients, clientID)
			s.mu.Unlock()
			return err
		}
		data := []byte(msg.Text)
		auth, err := s.authenticator.Authenticate(msg.Username, msg.Signature, data)
		if err != nil || !auth {
			fmt.Println(err)
			log.Println("error authenticating user")
		}

		s.mu.Lock()
		for _, client := range s.clients {
			client.Send(&message.SendMessageResponse{Text: fmt.Sprintf("%s: %s", msg.Username, msg.GetText())})
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
	auth := internal.NewInMemoryAuthenticator("/home/neal/workspace/shitchat/scratch/keys/")
	message.RegisterMessageServiceServer(s, &server{clients: make(map[string]message.MessageService_BroadcastServer), authenticator: auth})

	log.Println("Server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
