package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func messageSendingLoop(client proto.ChitChatClient) {
	var content string // the string to contain the message to be sent
	for {
		fmt.Scanln(&content)                                                    // capture input from the client
		client.PostMessage(context.Background(), &proto.ChatMsg{Text: content}) // send the message
	}
}

func messageReceivingLoop(stream grpc.ServerStreamingClient[proto.ChatMsg]) {
	for {
		msg, err := stream.Recv() // get some message
		if msg != nil {
			fmt.Println(msg.Text) // if this trulu was a message, print it out
		}
		if err == nil { // this does naffin for now
		}
		time.Sleep(time.Millisecond * 100) // check for message 10 times per second
	}
}

func main() {
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

	// keep two seperate loops for sending and receiving messages
	// putting both into the same message loop proved cumbersome
	go messageSendingLoop(client)
	go messageReceivingLoop(stream)
	for { // third loop for connection lifetime - possibly disconnecting should happen via a specialised message
	}

}
