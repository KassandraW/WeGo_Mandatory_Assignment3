package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Lamport clock
var lamportClock int32 = 0
var lamportLock sync.Mutex

func messageSendingLoop(client proto.ChitChatClient, name string) {
	var content string // the string to contain the message to be sent
	for {
		fmt.Scanln(&content)                                                                  // capture input from the client
		client.PostMessage(context.Background(), &proto.ChatMsg{Text: content, Sender: name}) // send the message
	}
}

func messageReceivingLoop(stream grpc.ServerStreamingClient[proto.ChatMsg], running *bool) {
	for {
		msg, err := stream.Recv() // get some message
		if msg != nil {
			lamportLock.Lock()
			lamportClock = max(lamportClock, msg.Timestamp) + 1                                 // sync the lamport clock
			log.Println(strconv.Itoa(int(lamportClock)) + " : " + msg.Sender + ": " + msg.Text) // if this truly was a message, print it out
			lamportLock.Unlock()
		}
		if err != nil {
			*running = false
		}
		time.Sleep(time.Millisecond * 100) // check for message 10 times per second
	}
}

func main() {
	// connect to server at localhost:5050. Insecure.newcredentials is used to skip TLS encryption for simplification
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	// create a client that can call RPC methods defined in the ChitChat service.
	client := proto.NewChitChatClient(conn)

	// establish connection stream to receive messages from server
	// step 1. increment lamport clock
	lamportLock.Lock()
	lamportClock += 1

	// step 2. Join the server
	stream, err := client.Join(context.Background(), &proto.Timestamp{Timestamp: lamportClock})
	if err != nil {
		log.Fatalf("Connection was not established")
	}
	// we are done using the lamportclock - unlock it
	lamportLock.Unlock()

	// receive our name for the session
	name_msg, err := stream.Recv()
	if err != nil {
		log.Fatalf("Failed to receive message")
	}
	lamportLock.Lock()
	lamportClock = max(lamportClock, name_msg.Timestamp) + 1

	name := name_msg.Text
	if name == "no names left" {
		fmt.Println(strconv.Itoa(int(lamportClock)) + " : the chat room is currently full! try again later.")
	} else {
		fmt.Println(strconv.Itoa(int(lamportClock)) + " : welcome to the chatroom! your name for this session is: " + name)
	}
	lamportLock.Unlock()

	running := true

	// keep two seperate loops for sending and receiving messages
	// putting both into the same message loop proved cumbersome
	go messageSendingLoop(client, name)
	go messageReceivingLoop(stream, &running)
	for running { // third loop for connection lifetime - possibly disconnecting should happen via a specialised message
	}

}
