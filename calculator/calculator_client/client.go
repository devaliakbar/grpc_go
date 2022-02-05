package main

import (
	"calculator/calculatorpb"
	"context"

	"io"
	"log"
	"time"

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
	//doServerStream(c)
	doClientStream(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	log.Println("Client Requesting")

	req := calculatorpb.CalculatorRequest{
		Number1: 3,
		Number2: 2,
	}

	res, err := c.Calculator(context.Background(), &req)
	if err != nil {
		log.Fatalf("Failed to calculate: %s", err)
	}

	log.Printf("Response from server :%s\n", res)
}

func doServerStream(c calculatorpb.CalculatorServiceClient) {
	log.Println("Get prime client Requesting")

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

func doClientStream(c calculatorpb.CalculatorServiceClient) {
	log.Println("Get Average client Requesting")

	requests := []*calculatorpb.GetAverageRequest{
		{
			Number: 1,
		},
		{
			Number: 2,
		},
		{
			Number: 3,
		},
		{
			Number: 4,
		},
	}

	stream, err := c.GetAverage(context.Background())
	if err != nil {
		log.Fatalf("Failed to request: %s", err)
	}

	for _, req := range requests {
		log.Printf("Sending request: %s", req)
		stream.Send(req)
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to get average: %s", err)
	}

	log.Printf("Average : %s", res)
}
