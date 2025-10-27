package main

import (
	proto "ChitChatServer/grpc"
	"fmt"

	"context"
	"log"
	"os"
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

	// creating a seperate log file : used this guide : https://last9.io/blog/write-logs-to-file/
	log_File, err := os.OpenFile("Log_info", os.O_CREATE |os.O_WRONLY| os.O_APPEND,0666) // create file if not exist|open file for writing | new issue goes to button no overwriting 
		if (err != nil){
			log.Fatal("could not open log file:%v", err)
		}
		defer log_File.Close()

	log.SetOutput(log_File)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("does this go to file 1?")

	fmt.Println("ChitChat Server is up and runnning.")
	server.start_server() //starts the gRPC server
	log.Printf("does this go to file 2?")
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

