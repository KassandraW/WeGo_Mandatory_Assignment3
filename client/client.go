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

func messageWritingLoop(client proto.ChitChatClient) {
	var content string
	for {
		fmt.Scanln(&content)
		client.PostMessage(context.Background(), &proto.ChatMsg{Text: content})
	}
}

func messageReceivingLoop(stream grpc.ServerStreamingClient[proto.ChatMsg]) {
	for {
		msg, err := stream.Recv()
		if msg != nil {
			fmt.Println(msg.Text)
		}
		if err == nil {
		}
		time.Sleep(time.Millisecond * 100)
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

	go messageWritingLoop(client)
	go messageReceivingLoop(stream)
	for { // connection lifetime loop
	}

}
