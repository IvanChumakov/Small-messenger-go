package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "chat/protos"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMessagingServiceServer
	subscribers map[string]chan *pb.MessageResponse
	mu          sync.Mutex
}

func NewServer() *server {
	return &server{
		subscribers: make(map[string]chan *pb.MessageResponse),
	}
}

func (s *server) SendMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[req.RecipientId]; ok {
		msg := &pb.MessageResponse{
			SenderId: req.RecipientId,
			Message:  req.Message,
		}
		ch <- msg
		return msg, nil
	}

	return nil, nil
}

func (s *server) Subscribe(req *pb.SubscriptionRequest, stream pb.MessagingService_SubscribeServer) error {
	ch := make(chan *pb.MessageResponse)

	s.mu.Lock()
	s.subscribers[req.ClientId] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.subscribers, req.ClientId)
	}()

	for msg := range ch {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessagingServiceServer(grpcServer, NewServer())

	log.Println("Starting server on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
