package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/devaliakbar/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	greetpb.UnimplementedGreetServiceServer
}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Server hitted :%v\n", req)

	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	res := &greetpb.GreetResponse{
		Result: result,
	}

	return res, nil
}

func (s *server) GreetDeadline(ctx context.Context, req *greetpb.GreetDeadlineRequest) (*greetpb.GreetDeadlineResponse, error) {
	fmt.Printf("Greet deadline hitted :%v\n", req)

	time.Sleep(3 * time.Second)

	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	res := &greetpb.GreetDeadlineResponse{
		Result: result,
	}

	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("The client canceled the request!")
			return nil, status.Error(codes.Canceled, "the client canceled the request")
		}

		return nil, status.Errorf(
			codes.Aborted,
			"Unexpected error",
		)
	}
	return res, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("Stream response: %s", req)

	firstName := req.GetGreeting().GetFirstName()

	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " : " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{Response: result}
		stream.Send(res)
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("Stream requested")

	result := ""

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Request stream done")
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Failed to listen error: %s", err)
		}

		result += req.GetGreeting().GetFirstName() + "! "
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("Bi Stream")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Request stream done")
			return nil
		}

		if err != nil {
			log.Fatalf("Failed to listen error: %s", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName

		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if sendErr != nil {
			log.Fatalf("Failed to send: %s", sendErr)
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	tls := false
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed loading certificate")
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)

	greetpb.RegisterGreetServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
