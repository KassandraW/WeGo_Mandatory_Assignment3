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

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials())) //connects to server at localhost:5050. Insecure.newcredentials is used to skip TLS encryption for simplification
	if err != nil {
		log.Fatalf("Not working")
	}

	client := proto.NewChitChatClient(conn) //creates a client that can call RPC methods defined in the ChitChat service.
	// Establish connection stream to recaive messages from server
	connection, err := client.ServerStream(context.Background(), &proto.Chat_Request{Greeting: "sup"})
	client.PostMessage(connection.Context(), &proto.ChatMsg{Text: "hello guys"})
	if err != nil {
		log.Fatalf("Connection was not established")
	}
	for {
		msg, err := connection.Recv()
		if msg != nil {
			fmt.Println(msg.Text)
		}
		if err == nil {
		}
		time.Sleep(time.Millisecond * 100)

		//client.PostMessage(context.Background(), &proto.ChatMsg{Text: "hey guys"})
	}

}
