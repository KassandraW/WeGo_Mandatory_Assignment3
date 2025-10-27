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
	stream proto.ChitChat_JoinServer
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
	// Copy the client list while holding the lock, then send without the lock.
	// This avoids deadlocks when Broadcast is called from code that already
	// holds s.lock.
	s.lock.Lock()
	clientsCopy := make([]ClientWrapper, len(s.clients))
	copy(clientsCopy, s.clients)
	s.lock.Unlock()

	for _, client := range clientsCopy {
		// Send may block; do it without holding the server lock.
		client.stream.Send(msg)
	}
}

func (s *ChitChatServer) PostMessage(ctx context.Context, msg *proto.ChatMsg) (*proto.Empty, error) {
	s.Broadcast(msg)
	s.lock.Lock()
	s.syncClock(msg.Timestamp)
	log.Println("Server ; T = " + strconv.Itoa(int(s.lamportClock)) + " ; Broadcast ; \"" + msg.Text + "\" ; Timestamp = " + strconv.Itoa(int(msg.Timestamp)))
	s.lock.Unlock()
	return &proto.Empty{}, nil
}

func (s *ChitChatServer) RemoveClient(target ClientWrapper) {
	for i, client := range s.clients {
		if client == target {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
		}
	}
}

func (s *ChitChatServer) Join(timestamp *proto.Timestamp, stream proto.ChitChat_JoinServer) error {
	// Gather state under lock, then perform Send/Broadcast without holding the lock.
	s.lock.Lock()
	s.syncClock(timestamp.Timestamp)
	lamportBefore := s.lamportClock // timestamp used for the name message

	// give the client a name and determine if they can join
	name := s.chooseRandomName()
	new_client := ClientWrapper{name: name, stream: stream}

	if name == "no names left" {
		// reply with the current logical time and exit
		s.lock.Unlock()
		stream.Send(&proto.ChatMsg{
			Text:      name,
			Sender:    "Server",
			Timestamp: lamportBefore})
		return nil
	}

	// add client and prepare broadcast message/timestamp
	s.clients = append(s.clients, new_client)
	s.lamportClock += 1
	broadcastTS := s.lamportClock
	broadcastMsg := &proto.ChatMsg{
		Text:      "Participant " + name + " joined Chit Chat at logical time " + strconv.Itoa(int(broadcastTS)),
		Sender:    "Server",
		Timestamp: broadcastTS,
	}
	s.lock.Unlock()

	// Send the assigned name to the joining client (uses the pre-increment timestamp)
	stream.Send(&proto.ChatMsg{
		Text:      name,
		Sender:    "Server",
		Timestamp: lamportBefore})
	log.Println("Sent name to joining : \"" + name + "\" at logical time " + strconv.Itoa(int(lamportBefore)))

	// Log and broadcast the join (does not hold the server lock)
	log.Println("Participant " + name + " joined Chit Chat at logical time " + strconv.Itoa(int(broadcastTS)))
	s.Broadcast(broadcastMsg)

	//keep the stream open until disconnection
	for {
		select {
		case <-stream.Context().Done(): //handle the client disconnecting
			s.lock.Lock()
			// handle removing the client and prepare broadcast while holding the lock
			s.RemoveClient(new_client)
			s.recycleName(name)
			s.lamportClock += 1
			leftTS := s.lamportClock
			leftMsg := &proto.ChatMsg{
				Text:      "Participant " + name + " left Chit Chat at logical time " + strconv.Itoa(int(leftTS)),
				Sender:    "Server",
				Timestamp: leftTS}
			s.lock.Unlock()

			// log and broadcast without holding the lock
			log.Println("Participant " + name + " left Chit Chat at logical time " + strconv.Itoa(int(leftTS)))
			s.Broadcast(leftMsg)
			return nil

		default:
			time.Sleep(500 * time.Millisecond) //wait half a second before checking again
		}
	}
}

func (s *ChitChatServer) syncClock(clientClock int32) {
	if clientClock > s.lamportClock {
		s.lamportClock = clientClock
	}
	s.lamportClock += 1
}

func main() { //initializes server
	var anonymous_client_names = []string{"Mercy", "Ana", "Lucio", "Reinhardt", "Roadhog", "Sigma", "Soldier 76", "Ashe", "Sombra"}
	server := &ChitChatServer{anonymous_client_names: anonymous_client_names}
	server.lamportClock = 0
	server.lock.Lock()
	server.lamportClock += 1
	log.Println("The ChitChat Server is now up and runnning at logical time " + strconv.Itoa(int(server.lamportClock)))
	server.lock.Unlock()
	server.start_server() //starts the gRPC server
}

func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()              //Creates a new gRPC server instance
	listener, err := net.Listen("tcp", ":5050") //Listens on TCP port 5050 using net.Listen
	if err != nil {
		log.Fatalf("Did not work")
	}
	proto.RegisterChitChatServer(grpcServer, s) //registers the server implementation with gRPC
	go func() {
		for {
			var content string
			fmt.Scanln(&content)
			if content == "exit" {
				s.lock.Lock()
				s.lamportClock += 1
				log.Println("The ChitChat Server is shutting down at logical time " + strconv.Itoa(int(s.lamportClock)))
				fmt.Println("Please enter Ctrl + C to completely shut down.")
				shutdown_message := &proto.ChatMsg{
					Text:      "The ChitChat Server is shutting down at logical time " + strconv.Itoa(int(s.lamportClock)),
					Sender:    "server",
					Timestamp: s.lamportClock,
				}
				s.Broadcast(shutdown_message)
				s.lock.Unlock()
				time.Sleep(3 * time.Second)
				grpcServer.GracefulStop() //the server is now closed for rpc calls
				listener.Close()
				return
			}
		}
	}()
	err = grpcServer.Serve(listener) // starts serving incoming requests
	if err != nil {
		log.Fatalf("Did not work")
	}
}
