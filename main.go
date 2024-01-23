package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/adityachandla/graph_access_service/generated"
	"github.com/adityachandla/graph_access_service/graphaccess"
	"github.com/adityachandla/graph_access_service/storage"
	"google.golang.org/grpc"
)

//go:generate protoc --go-grpc_out=generated --go_out=generated --go_opt=paths=source_relative  --go-grpc_opt=paths=source_relative graph_access.proto
var (
	port   = flag.Int("port", 20301, "The server port")
	fsType = flag.String("fstype", "s3", "Filesystem type s3/local")
	bucket = flag.String("bucket", "s3graphtest1", "Path to the s3 bucket")
	noLog  = flag.Bool("nolog", false, "Turn off logging")
	region = flag.String("region", "eu-west-1", "AWS Region")
)

type server struct {
	pb.UnimplementedGraphAccessServer
	accessService graphaccess.GraphAccess
}

func (s *server) GetNeighbours(ctx context.Context, req *pb.AccessRequest) (*pb.AccessResponse, error) {
	log.Printf("Processing reqest %v\n", req)
	request := graphaccess.Request{
		Node:      req.NodeId,
		Label:     req.Label,
		Direction: mapDirection(req.Direction),
	}
	neighbours, err := s.accessService.GetNeighbours(request)
	response := &pb.AccessResponse{Neighbours: neighbours}
	if err != nil {
		response.Status = pb.AccessResponse_SERVER_ERROR
		return response, err
	}
	response.Status = pb.AccessResponse_NO_ERROR
	return response, nil
}

func mapDirection(dir pb.AccessRequest_Direction) graphaccess.Direction {
	if dir == pb.AccessRequest_OUTGOING {
		return graphaccess.OUTGOING
	} else if dir == pb.AccessRequest_INCOMING {
		return graphaccess.INCOMING
	}
	return graphaccess.BOTH
}

func main() {
	flag.Parse()
	if *noLog {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}
	var fetcher storage.Fetcher
	if *fsType == "s3" {
		fetcher = storage.InitializeS3Service(*bucket, *region)
	} else if *fsType == "local" {
		fetcher = storage.InitializeFsService(*bucket)
	} else {
		panic("Invalid filesystem type")
	}
	simple_csr := graphaccess.InitializeSimpleCsrAccess(fetcher)
	server := &server{accessService: simple_csr}
	log.Println("Initialized the server")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterGraphAccessServer(s, server)
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Unable to serve request: %v", err)
	}
}
