package main

import (
	"context"
	hw "github.com/ueno-bst/grpc-gateway-skel/examples/grpc/skelton/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
)

type helloService struct {
	hw.UnimplementedHelloWorldServiceServer
}

func (w *helloService) Header(ctx context.Context, req *hw.HeaderRequest) (*hw.HeaderResponse, error) {
	values := make(map[string]*hw.HeaderResponse_Value)

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for s, strings := range md {
			values[s] = &hw.HeaderResponse_Value{
				Value: strings,
			}
		}
	}

	return &hw.HeaderResponse{
		Headers: values,
	}, nil
}

func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:8080")

	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	hw.RegisterHelloWorldServiceServer(s, &helloService{})

	reflection.Register(s)

	go func() {
		log.Printf("start gRPC server port: %v", "0.0.0.0:8080")
		_ = s.Serve(listener)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}
