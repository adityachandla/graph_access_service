package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/adityachandla/graph_access_service/generated"
	"google.golang.org/grpc"
)

//go:generate protoc --go-grpc_out=generated --go_out=generated --go_opt=paths=source_relative  --go-grpc_opt=paths=source_relative graph_access.proto
var (
	port = flag.Int("port", 20301, "The server port")
)

type server struct {
	pb.UnimplementedGraphAccessServer
}

func (s *server) GetNeighbours(ctx context.Context,
	req *pb.AccessRequest) (*pb.AccessResponse, error) {
	response := &pb.AccessResponse{
		Neighbours: []uint32{1, 2, 3},
		Status:     pb.AccessResponse_NO_ERROR,
	}
	log.Printf("Served request")
	return response, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterGraphAccessServer(s, &server{})
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Unable to serve request: %v", err)
	}
}
