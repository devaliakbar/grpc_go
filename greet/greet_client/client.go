package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/devaliakbar/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {

	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())

	tls := false
	if tls {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}

		transportOption = grpc.WithTransportCredentials(tlsCredentials)
	}

	cc, err := grpc.Dial("0.0.0.0:50051", transportOption)

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	doUnary(c)
	//doDeathlineUnary(c, 4*time.Second) //Change this second 2 to see timeout
	//doServerStream(c)
	//doClientStream(c)
	//doBiStream(c)
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
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

func doDeathlineUnary(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Client requesting")

	req := greetpb.GreetDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ali",
			LastName:  "Akbar",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetDeadline(ctx, &req)
	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				log.Fatalf("Request timeout: %s", statusErr.Message())
			}
			log.Fatalf("Failed response from server: %s", statusErr.Message())
		}
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

func doClientStream(c greetpb.GreetServiceClient) {
	fmt.Println("Client streaming")

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Ali",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Akbar",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Ajay",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Arya",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Akshay",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Failed to LongGreet: %s", err)
	}

	for _, req := range requests {
		log.Printf("Requesting: %s", req)
		stream.Send(req)
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Failed to get Long request's response: %s", err)
	}

	log.Printf("Long request's response: %s", res)
}

func doBiStream(c greetpb.GreetServiceClient) {
	fmt.Println("Bi streaming")

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Failed to strem: %s", err)
	}

	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Ali",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Akbar",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Ajay",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Arya",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Akshay",
			},
		},
	}

	waitc := make(chan struct{})

	go func() {
		for _, req := range requests {
			log.Printf("Sending bi request: %s", req)
			stream.Send(req)
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("Failed to get stream:%s", err)
				break
			}

			log.Printf("Response from server: %s", res)
		}
		close(waitc)
	}()

	<-waitc
}
