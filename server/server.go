package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"google.golang.org/grpc"
)

// implements the gRPC service defined in the .proto file.
type ChitChatServer struct {
	proto.UnimplementedChitChatServer
	mu                     sync.Mutex
	clients                []ClientWrapper // slice of client streams for broadcasting
	anonymous_client_names []string
}

type ClientWrapper struct {
	name   string
	stream proto.ChitChat_ServerStreamServer
}


var anonymous_client_names = []string{"Mercy", "Ana", "Lucio", "Reinhart", "Roadhog", "Sigma", "Soldier 76", "Ashe", "Sombra"}

func (s *ChitChatServer) chooseRandomName() string {
	if len(s.anonymous_client_names) == 0 {
		return "no names left"
	}
	random_index := rand.Intn(len(s.anonymous_client_names))
	random_name := s.anonymous_client_names[random_index]
	s.anonymous_client_names = append(s.anonymous_client_names[:random_index], s.anonymous_client_names[random_index+1:]...)
	return random_name
}

func (s *ChitChatServer) recycleName(name string) {
	s.anonymous_client_names = append(s.anonymous_client_names, name)
}

func (s *ChitChatServer) Broadcast(msg *proto.ChatMsg) {
	s.mu.Lock() // chat gpt recommends doing this
	defer s.mu.Unlock()
	for _, client := range s.clients {
		client.stream.Send(msg)
	}
}

func (s *ChitChatServer) PostMessage(ctx context.Context, msg *proto.ChatMsg) (*proto.Empty, error) {
	s.Broadcast(msg)
	return &proto.Empty{}, nil
}

func (s *ChitChatServer) RemoveClient(target ClientWrapper) {
	for i, client := range s.clients {
		if client == target {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
		}
	}
}

func (s *ChitChatServer) ServerStream(request *proto.Chat_Request, stream proto.ChitChat_ServerStreamServer) error {
	//Keep track of clientID and send it to the client
	name := s.chooseRandomName()
	if name == "no names left" {
		stream.Send(&proto.ChatMsg{Text: name})
		return nil
	} else {
		stream.Send(&proto.ChatMsg{Text: name})
	}

	new_client := ClientWrapper{name: name, stream: stream}
	s.clients = append(s.clients, new_client) // Add the stream to list of streams

	s.Broadcast(&proto.ChatMsg{Text: name + " just joined!", Sender: "Server"}) // let them know the goat has arrived
	<-stream.Context().Done()                                                   // block until the goat leaves
	s.RemoveClient(new_client)                                                  // remove the stream from the list
	s.recycleName(name)
	s.Broadcast(&proto.ChatMsg{Text: name + " just left.", Sender: "Server"}) // let them know the goat has left
	return nil
}

func main() { //initializes server
filepath := "../grpc/Log_info"
	// creating a seperate log file : used this guide : https://last9.io/blog/write-logs-to-file/
 Log_File, err := os.OpenFile(filepath, os.O_CREATE |os.O_WRONLY| os.O_APPEND,0666) // create file if not exist|open file for writing | new issue goes to button no overwriting 
		if (err != nil){
			log.Fatal("could not open log file: %v", err)
		}
		defer Log_File.Close()

	log.SetOutput(Log_File)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	server := &ChitChatServer{anonymous_client_names: anonymous_client_names}
	fmt.Printf("ChitChat Server is up and runnning.\n")
	log.Print("server start started")
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



