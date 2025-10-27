package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials())) //connects to server at localhost:5050. Insecure.newcredentials is used to skip TLS encryption for simplification
	if err != nil {
		log.Fatalf("Not working")
	}
	log.Printf("Hello world !")

	client := proto.NewChitChatClient(conn) //creates a client that can call RPC methods defined in the ChitChat service.
	
	connectionstatus, err := client.GetConnection(context.Background(), &proto.Empty{}) //Client calls getstudents with an empty request and receives a students message containing the list of names.
	if err != nil {
		log.Fatalf("Not working")
	} 
	fmt.Print(connectionstatus)
}
