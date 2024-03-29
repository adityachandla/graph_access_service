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
	port     = flag.Int("port", 20301, "The server port")
	fsType   = flag.String("fstype", "s3", "Filesystem type s3/local")
	bucket   = flag.String("bucket", "s3graphtest1", "Path to the s3 bucket")
	noLog    = flag.Bool("nolog", false, "Turn off logging")
	region   = flag.String("region", "eu-west-1", "AWS Region")
	accessor = flag.String("accessor", "prefetch", "Possible values are: prefetch/offset/simple")
)

type server struct {
	pb.UnimplementedGraphAccessServer
	accessService graphaccess.GraphAccess
}

func (s *server) GetNeighbours(_ context.Context, req *pb.AccessRequest) (*pb.AccessResponse, error) {
	log.Printf("Processing reqest %v\n", req)
	request := graphaccess.Request{
		Node:      req.NodeId,
		Label:     req.Label,
		Direction: mapDirection(req.Direction),
	}
	response := &pb.AccessResponse{Neighbours: s.accessService.GetNeighbours(request)}
	response.Status = pb.AccessResponse_NO_ERROR
	return response, nil
}

func (s *server) GetStats(_ context.Context, _ *pb.StatsRequest) (*pb.Stats, error) {
	return &pb.Stats{Stats: s.accessService.GetStats()}, nil
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
	fetcher := getFetcher()
	accessService := getAccessService(fetcher)
	log.Println("Initialized access service")
	s := &server{accessService: accessService}
	startServer(s)
}

func startServer(ser *server) {
	//Server start
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterGraphAccessServer(s, ser)
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Unable to serve request: %v", err)
	}
}

func getFetcher() storage.Fetcher {
	if *fsType == "s3" {
		return storage.InitializeS3Service(*bucket, *region)
	} else if *fsType == "local" {
		return storage.InitializeFsService(*bucket)
	} else {
		panic("Invalid filesystem type")
	}
}

func getAccessService(fetcher storage.Fetcher) graphaccess.GraphAccess {
	if *accessor == "simple" {
		return graphaccess.NewSimpleCsr(fetcher)
	} else if *accessor == "offset" {
		return graphaccess.NewOffsetCsr(fetcher)
	} else if *accessor == "prefetch" {
		return graphaccess.NewPrefetchCsr(fetcher)
	} else {
		panic("Invalid accessor")
	}
}
