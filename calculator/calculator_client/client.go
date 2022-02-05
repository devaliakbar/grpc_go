package main

import (
	"calculator/calculatorpb"
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)
	//doUnary(c)
	doServerStream(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Client Requesting")

	req := calculatorpb.CalculatorRequest{
		Number1: 3,
		Number2: 2,
	}

	res, err := c.Calculator(context.Background(), &req)
	if err != nil {
		log.Fatalf("Failed to calculate: %s", err)
	}

	fmt.Printf("Response from server :%s\n", res)
}

func doServerStream(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Get prime client Requesting")

	req := calculatorpb.GetPrimeRequest{
		Number: 120,
	}

	resStream, err := c.GetPrime(context.Background(), &req)
	if err != nil {
		log.Fatalf("Failed to calculate: %s", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			log.Println("Completed")
			break
		}

		if err != nil {
			log.Fatalf("Error happened while listen: %s", err)
		}

		log.Printf("Response from server: %s", msg)
	}
}
