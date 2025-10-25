package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// implements the gRPC service defined in the .proto file.
type ChitChatServer struct {
	proto.UnimplementedChitChatServer
	clients                []ClientWrapper // slice of client streams for broadcasting
	anonymous_client_names []string
	lamportClock           int32
	lock                   sync.Mutex
}

type ClientWrapper struct {
	name   string
	stream proto.ChitChat_GetServerStreamServer
}

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
	s.lock.Lock()
	defer s.lock.Unlock() // ensures the lock unlocks even if Send panicks or if there was a return inside the loop
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

func (s *ChitChatServer) ServerStream(timestamp *proto.Timestamp, stream proto.ChitChat_GetServerStreamServer) error {
	// lamport clock business
	s.lock.Lock()
	s.syncClock(timestamp.Timestamp)
	lamportTimeStr := strconv.Itoa(int(s.lamportClock)) // for broadcasting purposes
	s.lock.Unlock()

	//keep track of client name and send it to the client
	name := s.chooseRandomName()
	if name == "no names left" {
		stream.Send(&proto.ChatMsg{Text: name, Sender: "server"})
		return nil
	} else {
		stream.Send(&proto.ChatMsg{Text: name, Sender: "server"})
	}

	// add the client and its stream to the list of clients
	s.lock.Lock()
	new_client := ClientWrapper{name: name, stream: stream}
	s.clients = append(s.clients, new_client)
	s.lock.Unlock()

	// log the client joining
	log.Println("Participant " + name + " joined Chit Chat at logical time" + lamportTimeStr)

	//broadcast the client joining
	s.Broadcast(&proto.ChatMsg{Text: "Participant " + name + " joined Chit Chat at logical time" + lamportTimeStr, Sender: "Server"})

	//keep the stream open until disconnection
	for {
		select {
		case <-stream.Context().Done(): //handle the client disconnecting
			s.lock.Lock()
			//handle removing the client
			s.RemoveClient(new_client)
			s.recycleName(name)

			//broadcast that the client has left
			s.lamportClock += 1
			s.Broadcast(&proto.ChatMsg{Text: "Participant " + name + " left Chit Chat at logical time " + strconv.Itoa(int(s.lamportClock)), Sender: "Server"})
			s.lock.Unlock()
			return nil

		default:
			time.Sleep(500 * time.Millisecond) //wait half a second before checking again
		}
	}
}

func (s *ChitChatServer) syncClock(clientClock int32) {
	if clientClock > s.lamportClock {
		s.lamportClock = clientClock
		s.lamportClock += 1
	}
}

func main() { //initializes server
	var anonymous_client_names = []string{"Mercy", "Ana", "Lucio", "Reinhart", "Roadhog", "Sigma", "Soldier 76", "Ashe", "Sombra"}
	server := &ChitChatServer{anonymous_client_names: anonymous_client_names}
	server.lamportClock = 0

	log.Println("The ChitChat Server is now up and runnning")
	fmt.Println("The ChitChat Server is now up and runnning")
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
