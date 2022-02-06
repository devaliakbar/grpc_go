package main

import (
	"chat/chatpb"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type chatServiceServer struct {
	chatpb.UnimplementedChatServiceServer
	channel map[string][]chan *chatpb.Message
}

func (s *chatServiceServer) JoinChannel(ch *chatpb.Channel, msgStream chatpb.ChatService_JoinChannelServer) error {

	msgChannel := make(chan *chatpb.Message)
	s.channel[ch.Name] = append(s.channel[ch.Name], msgChannel)

	// doing this never closes the stream
	for {
		select {
		case <-msgStream.Context().Done():
			return nil
		case msg := <-msgChannel:
			fmt.Printf("GO ROUTINE (got message): %v \n", msg)
			msgStream.Send(msg)
		}
	}
}

func (s *chatServiceServer) SendMessage(ctx context.Context, message *chatpb.Message) (*chatpb.MessageAck, error) {
	log.Printf("New message received in server: %s", message)

	ack := &chatpb.MessageAck{Status: "SENT"}

	go func() {
		streams := s.channel[message.Channel.Name]
		for _, msgChan := range streams {
			msgChan <- message
		}
	}()

	return ack, nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	chatpb.RegisterChatServiceServer(s, &chatServiceServer{
		channel: make(map[string][]chan *chatpb.Message),
	})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
