package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc"
)

// implements the gRPC service defined in the .proto file.
type ChitChatServer struct {
	proto.UnimplementedChitChatServer
	mu             sync.Mutex
	client_streams []proto.ChitChat_ServerStreamServer // slice of client streams for broadcasting
}

func (s *ChitChatServer) Broadcast(msg string) {
	s.mu.Lock() // chat gpt recommends doing this
	defer s.mu.Unlock()
	for _, stream := range s.client_streams {
		stream.Send(&proto.ChatMsg{Text: msg})
	}
}

func (s *ChitChatServer) PostMessage(ctx context.Context, msg *proto.ChatMsg) (*proto.Empty, error) {
	s.Broadcast(msg.Text)
	return &proto.Empty{}, nil
}

func (s *ChitChatServer) RemoveStream(target proto.ChitChat_ServerStreamServer) {
	for i, stream := range s.client_streams {
		if stream == target {
			s.client_streams = append(s.client_streams[:i], s.client_streams[i+1:]...)
		}
	}
}

func (s *ChitChatServer) ServerStream(request *proto.Chat_Request, stream proto.ChitChat_ServerStreamServer) error {
	s.client_streams = append(s.client_streams, stream) // Add the stream to list of streams

	//Keep track of clientID and send it to the client
	clientId := len(s.client_streams)
	s.Broadcast(strconv.Itoa(clientId) + " just joined!") // let them know the goat has arrived
	<-stream.Context().Done()                             // block until the goat leaves
	s.RemoveStream(stream)                                // remove the stream from the list
	s.Broadcast(strconv.Itoa(clientId) + " just left.")   // let them know the goat has left
	return nil
}

func main() { //initializes server
	server := &ChitChatServer{}
	fmt.Printf("ChitChat Server is up and runnning.\n")
	server.start_server() //starts the gRPC server
}

func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()              //Creates a new gRPC server instance
	listener, err := net.Listen("tcp", ":5050") //Listens on TCP port 5050 using net.Listen
	if err != nil {
		log.Fatalf("Did not work")
	}
	proto.RegisterChitChatServer(grpcServer, s) //registers the server implementation with gRPC
	err = grpcServer.Serve(listener)            // starts serving incoming requests
	if err != nil {
		log.Fatalf("Did not work")
	}
}
