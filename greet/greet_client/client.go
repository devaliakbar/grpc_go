package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/devaliakbar/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	//doUnary(c)
	doServerStream(c)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Client requesting")

	req := greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ali",
			LastName:  "Akbar",
		},
	}

	res, err := c.Greet(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error greeting: %v", err)
	}

	fmt.Printf("Response from server: %v\n", res)
}

func doServerStream(c greetpb.GreetServiceClient) {
	fmt.Println("Client requesting server stream")

	req := greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ali",
			LastName:  "Akbar",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error greeting by server streming: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			log.Println("Completed!!!")
			break
		}

		if err != nil {
			log.Fatalf("Error happend while server streaming: %s", err)
		}

		log.Printf("Response from server: %s", msg)
	}

}
