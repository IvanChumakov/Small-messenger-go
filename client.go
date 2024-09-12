package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"

	pb "chat/protos"
	"google.golang.org/grpc"
)

func init() {
	pflag.StringP("id", "i", "client1", "User id")
	pflag.StringP("receiver", "r", "client2", "Receiver id")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewMessagingServiceClient(conn)
	clientID := viper.GetString("id")

	go func() {
		stream, err := client.Subscribe(context.Background(), &pb.SubscriptionRequest{ClientId: clientID})
		if err != nil {
			log.Fatalf("could not subscribe: %v", err)
		}
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Fatalf("error receiving message: %v", err)
			}
			fmt.Printf("%s: %s", msg.SenderId, msg.Message)
		}
	}()

	time.Sleep(time.Second)

	go func() {
		for {
			in := bufio.NewReader(os.Stdin)
			message, err := in.ReadString('\n')

			_, err = client.SendMessage(context.Background(), &pb.MessageRequest{RecipientId: viper.GetString("receiver"), Message: message})
			if err != nil {
				log.Fatalf("could not send message: %v", err)
			}
		}
	}()

	select {}
}
