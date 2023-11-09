package main

import (
	"bufio"
	pb "chat/Chat/Chat"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "localhost:7000", "Server address")
)

func main() {
	flag.Parse()
	var clientId string = ""
	fmt.Println("--- CLIENT APP ---")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	// fmt.Println("Address : ", *addr)
	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		log.Fatalf("Fail to dail: %v", err)
	} else {
		fmt.Println("Connection Established")
	}
	reader := bufio.NewReader(os.Stdin)
	defer conn.Close()
	fmt.Print("Enter your name : ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error in reading name from console ", err)
		return
	}
	fmt.Print("Enter your email : ")
	email, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error in reading name from console ", err)
		return
	}
	ctx := context.Background()
	client := pb.NewMessageClient(conn)

	// fmt.Println("Client is Created")
	response, err := client.Register(ctx, &pb.RegisterRequest{
		Username: name,
		Email:    email,
	})
	if err != nil {
		log.Println("error occurs in register method ", err)
		return
	}
	// fmt.Println("response : ",response)
	clientId = response.GetUserId()
	fmt.Println("Client ID is : ", clientId)
	go receiveMessages(ctx,client,clientId);
	for{
		fmt.Print("Enter Reciever name : ")
		reciever, err := reader.ReadString('\n')
		if(err!=nil){
			log.Fatalf("Error Occurs while reading reciever")
		}
		rec := strings.TrimSpace(reciever);
		// fmt.Println("Message", reciever)
		fmt.Print("Enter Message : ")
		message, err := reader.ReadString('\n')
		msg := strings.TrimSpace(message)
		go sendMessage(ctx, client, rec, msg, name)
	}
}

func receiveMessages(ctx context.Context, client pb.MessageClient, clientId string) {
	// fmt.Println("Receive method called")
	waitc := make(chan struct{})
	response, err := client.RecieveMessage(ctx, &pb.RecieveMessageRequest{
		ClientID: clientId,
	})
	if err != nil {
		fmt.Println("Error occurs in rpc call of recieveMessage Method")
		return
	}

	go func() {
		for {
			// fmt.Println("Inside the loop")
			in, err := response.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nErr: %v", err)
			}
			if in.Sender == "Server" {
				fmt.Printf("MESSAGE: (%v) -> %v \n", in.Sender, in.Text)
				close(waitc)
				os.Exit(1)
			}
			fmt.Printf("MESSAGE: (%v) -> %v \n", in.Sender, in.Text)
			// fmt.Println("\n MESSAGE : ( ",in.Sender,") -> ",in.Text)
		}
	}()
	<-waitc
}

func sendMessage(ctx context.Context, client pb.MessageClient, Reciever string, Text string, Sender string) {
	stream, err := client.SendMessage(ctx)
	if err != nil {
		fmt.Println("Error occurs in SendMessage method ", err)
		return
	}
	msg := pb.SendMessageRequest{
		Reciever: Reciever,
		Sender:   Sender,
		Text:     Text,
	}
	stream.Send(&msg)
	_, ok := stream.CloseAndRecv()
	if ok != nil {
		fmt.Println("Error in Send Message : ", err)
	}
	// fmt.Printf("Message sent: %v \n", ack.GetStatus());
}
