package main

import (
	"calculator/calculatorpb"
	"context"
	"fmt"
	"math"

	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

func (*server) Calculator(ctx context.Context, req *calculatorpb.CalculatorRequest) (*calculatorpb.CalculatorResponse, error) {
	log.Printf("Server hitted: %s", req)

	num1 := req.GetNumber1()
	num2 := req.GetNumber2()

	res := &calculatorpb.CalculatorResponse{
		Result: num1 + num2,
	}

	return res, nil
}

func (*server) GetPrime(req *calculatorpb.GetPrimeRequest, stream calculatorpb.CalculatorService_GetPrimeServer) error {
	log.Printf("Server prime hitted: %s", req)

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

func (*server) GetAverage(stream calculatorpb.CalculatorService_GetAverageServer) error {
	log.Printf("Server average hitted")

	var total int32
	var count int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Request stream end")
			return stream.SendAndClose(&calculatorpb.GetAverageResponse{
				Response: float64(total) / float64(count),
			})
		}

		if err != nil {
			log.Fatalf("Failed to listen request: %s", err)
		}

		total += req.GetNumber()
		count++
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	log.Printf("Received squareroot RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Number is less than 0: %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
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
