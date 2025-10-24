package main

import (
	proto "ChitChatServer/grpc"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

//implements the gRPC service defined in the .proto file.
type ChitChatServer struct { 
	proto.UnimplementedChitChatServer 
	client_streams []proto.ChitChat_ServerStreamServer // slice of client streams for broadcasting
}

func (s *ChitChatServer) Broadcast(msg string){

}

func (s *ChitChatServer) ServerStream( request *proto.Chat_Request, stream proto.ChitChat_ServerStreamServer) error {
	s.client_streams = append(s.client_streams, stream); // Add the stream to list of streams

	//Keep track of clientID and send it to the client
	clientId := len(s.client_streams);
	stream.Send(&proto.ChatMsg{
		Text: strconv.Itoa(clientId),
	})

	//Log the client joining
	log.Printf("Participant " + strconv.Itoa(clientId) + " joined Chit Chat at logical time L\n")
	
	//Broadcast "client joined" message
	joinmsg := fmt.Sprintf("Participant %d joined Chit Chat at logical time L\n", clientId)
	s.Broadcast(joinmsg)

	//keep the stream alive until the client disconnects
	for{
		select{
		case <-stream.Context().Done(): //client disconnected
			log.Printf("Participant " + strconv.Itoa(clientId) + " left Chit Chat at logical time L", clientId)
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func main() { //initializes server
	server := &ChitChatServer{}
	fmt.Printf("ChitChat Server is up and runnning.\n")
	server.start_server() //starts the gRPC server
}

func (s *ChitChatServer) start_server() {
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

