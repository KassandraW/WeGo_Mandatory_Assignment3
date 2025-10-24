package main

import (
	proto "ChitChatServer/grpc"
	"fmt"

	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type ChitChat_Server struct { //implements the gRPC service efinted in the .proto file.
	proto.UnimplementedChitChatServer //Embeds UnimplementedITUDatabaseServer to satisfy the interface
}

func (s *ChitChat_Server) GetConnection(ctx context.Context, in *proto.Empty) (*proto.Connected, error) { //takes a context and an empty request
	return &proto.Connected{Connected: "yea"}, nil //returns a Students message containing the list of student names and a possible error.
}

func main() { //initializes server
	server := &ChitChat_Server{}
	fmt.Printf("ChitChat Server is up and runnning.")
	server.start_server() //starts the gRPC server
}

func (s *ChitChat_Server) start_server() {
	grpcServer := grpc.NewServer() //Creates a new gRPC server instance
	listener, err := net.Listen("tcp", ":5050") //Listens on TCP port 5050 using net.Listen
	if err != nil {
		log.Fatalf("Did not work")
	}
	proto.RegisterChitChatServer(grpcServer, s) //registers the server implementation with gRPC
	err = grpcServer.Serve(listener) // starts serving incoming requests
	if err != nil {
		log.Fatalf("Did not work")
	}
}
