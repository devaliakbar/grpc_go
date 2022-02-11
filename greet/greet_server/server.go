package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
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
			statusErr, ok := status.FromError(err)
			if ok {
				if statusErr.Code() == codes.Canceled {
					return status.Errorf(
						codes.Canceled,
						fmt.Sprintf("Request canceled: %s", statusErr.Message()),
					)
				}
				log.Fatalf("Failed to listen error: %s", statusErr.Message())
			}

			log.Fatalf("Failed to listen error: %s", err)
		}
		log.Printf("Client stream: %s", req.GetGreeting().GetFirstName())
		result += req.GetGreeting().GetFirstName() + "! "
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("Bi Stream")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Request stream done")
			return nil
		}

		if err != nil {
			statusErr, ok := status.FromError(err)
			if ok {
				if statusErr.Code() == codes.Canceled {
					log.Println("Request stream canceled")
					return status.Errorf(
						codes.Canceled,
						fmt.Sprintf("Request canceled: %s", statusErr.Message()),
					)
				}
				log.Fatalf("Failed to listen error: %s", statusErr.Message())
			}

			log.Fatalf("Failed to listen error: %s", err)
		}

		log.Printf("Recieved :%s", req)
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

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	const serverCertFile = "cert/server-cert.pem"
	const serverKeyFile = "cert/server-key.pem"
	const clientCACertFile = "cert/ca-cert.pem"

	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile(clientCACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	tls := false
	if tls {

		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatalf("cannot load TLS credentials: %s", err)
		}

		opts = append(opts, grpc.Creds(tlsCredentials))
	}

	s := grpc.NewServer(opts...)

	greetpb.RegisterGreetServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
