package mtsrv

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func GrpcServer(port string, fn func(*grpc.Server)) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// call the registeration function
	fn(s)

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return fmt.Errorf("returned from listener")
}
