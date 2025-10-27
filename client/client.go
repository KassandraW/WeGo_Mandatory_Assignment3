package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Lamport clock
var lamportClock int32 = 0
var lamportLock sync.Mutex

func log_event(event_type string, msg *proto.ChatMsg) {
	log.Println(msg.Sender + " ; T = " + strconv.Itoa(int(lamportClock)) + " ; " + event_type + " ; \"" + msg.Text + "\" ; Timestamp = " + strconv.Itoa(int(msg.Timestamp)))
}

func messageSendingLoop(client proto.ChitChatClient, name string) {
	var content string // the string to contain the message to be sent
	for {
		fmt.Scanln(&content) // capture input from the client
		lamportLock.Lock()   //update clock
		lamportClock += 1
		message := &proto.ChatMsg{Text: content, Sender: name, Timestamp: lamportClock}
		client.PostMessage(context.Background(), message) // send the message
		log_event("Send", message)
		lamportLock.Unlock()
	}
}

func messageReceivingLoop(stream grpc.ServerStreamingClient[proto.ChatMsg], running *bool) {
	for {
		msg, err := stream.Recv() // check for a message
		if msg != nil {
			lamportLock.Lock()
			lamportClock = max(lamportClock, msg.Timestamp) + 1 // sync the lamport clock
			log_event("Recv", msg)
			lamportLock.Unlock()
		}
		if err != nil {
			*running = false
		}
		time.Sleep(time.Millisecond * 100) // check for message 10 times per second
	}
}

func main() {
	filepath := "../grpc/Log_info"
	Log_File, err := os.OpenFile(filepath, os.O_CREATE |os.O_WRONLY| os.O_APPEND,0666) // create file if not exist|open file for writing | new issue goes to button no overwriting 
		if (err != nil){
			log.Fatal("could not open log file client: %v", err)
		}
		defer Log_File.Close()
			
	log.SetOutput(Log_File)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials())) //connects to server at localhost:5050. Insecure.newcredentials is used to skip TLS encryption for simplification
	// connect to server at localhost:5050. Insecure.newcredentials is used to skip TLS encryption for simplification
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}
	
	

	// create a client that can call RPC methods defined in the ChitChat service.
	client := proto.NewChitChatClient(conn)

	// --- Join the chat ---
	// step 1. increment lamport clock
	lamportLock.Lock()
	lamportClock += 1

	// step 2. Join the server
	stream, err := client.Join(context.Background(), &proto.Timestamp{Timestamp: lamportClock})
	if err != nil {
		log.Fatalf("Connection was not established")
	}

	// receive our name for the session
	name_msg, err := stream.Recv()
	if err != nil {
		log.Fatalf("Failed to receive message")
	}

	//update lamport clock
	lamportClock = max(lamportClock, name_msg.Timestamp) + 1

	name := name_msg.Text
	if name == "no names left" {
		log.Println(strconv.Itoa(int(lamportClock)) + " : Unknown client failed to join the chat. Chatroom is full.")
		return
	}
	log_event("RecvName", name_msg)
	lamportLock.Unlock()

	running := true

	// keep two seperate loops for sending and receiving messages
	// putting both into the same message loop proved cumbersome
	go messageSendingLoop(client, name)
	go messageReceivingLoop(stream, &running)
	for running { // third loop for connection lifetime - possibly disconnecting should happen via a specialised message
	}

}
