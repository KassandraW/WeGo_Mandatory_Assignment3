package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
			fmt.Println(msg.Sender + ": " + msg.Text) // if this truly was a message, print it out
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
	if err != nil {
		log.Fatalf("Not working")
	}
	
	

	client := proto.NewChitChatClient(conn) //creates a client that can call RPC methods defined in the ChitChat service.
	// Establish connection stream to recaive messages from server
	stream, err := client.ServerStream(context.Background(), &proto.Chat_Request{Greeting: "sup"})
	if err != nil {
		log.Fatalf("Connection was not established")
	}
	name_msg, err := stream.Recv()
	name := name_msg.Text
	if name == "no names left" {
		fmt.Println("the chat room is currently full! try again later.")
	} else {
		fmt.Println("welcome to the chatroom! your name for this session is: " + name)
		log.Println("welcome to the chatroom! your name for this session is: " + name)
	}

	running := true

	// keep two seperate loops for sending and receiving messages
	// putting both into the same message loop proved cumbersome
	go messageSendingLoop(client, name)
	go messageReceivingLoop(stream, &running)
	for running { // third loop for connection lifetime - possibly disconnecting should happen via a specialised message
	}

}
