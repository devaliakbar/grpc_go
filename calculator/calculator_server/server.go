package main

import (
	"calculator/calculatorpb"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

type server struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

func (*server) Calculator(ctx context.Context, req *calculatorpb.CalculatorRequest) (*calculatorpb.CalculatorResponse, error) {
	fmt.Printf("Server hitted: %s", req)

	num1 := req.GetNumber1()
	num2 := req.GetNumber2()

	res := &calculatorpb.CalculatorResponse{
		Result: num1 + num2,
	}

	return res, nil
}

func (*server) GetPrime(req *calculatorpb.GetPrimeRequest, stream calculatorpb.CalculatorService_GetPrimeServer) error {
	fmt.Printf("Server prime hitted: %s", req)

	number := req.GetNumber()

	var k int32 = 2

	for number > 1 {
		if number%k == 0 {
			res := &calculatorpb.GetPrimeResponse{Result: k}
			stream.Send(res)
			time.Sleep(1 * time.Second)
			number = number / k
		} else {
			k = k + 1
		}
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}

	s := grpc.NewServer()

	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
}
