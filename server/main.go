package main

import (
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/lukasmalkmus/stream/bytestream"
	pb "github.com/lukasmalkmus/stream/proto"
)

type bytestreamSrv struct{}

// Read is used to retrieve the contents of a resource as a sequence of bytes.
// The bytes are returned in a sequence of responses, and the responses are
// delivered as the results of a server-side streaming RPC.
func (b bytestreamSrv) Read(req *pb.ReadRequest, stream pb.ByteStream_ReadServer) error {
	if req.ResourceName == "" {
		return grpc.Errorf(codes.InvalidArgument, "ReadRequest: empty or missing resource_name")
	}

	f, err := os.Open(req.ResourceName)
	if err != nil {
		return grpc.Errorf(codes.NotFound, "ReadRequest: resource %q doesn't exist", req.ResourceName)
	}
	defer f.Close()

	buf := make([]byte, bytestream.MaxBufSize)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if err := stream.Send(&pb.ReadResponse{Data: buf[:n]}); err != nil {
				return grpc.Errorf(grpc.Code(err), "Send(resourceName=%q): %v", req.ResourceName, grpc.ErrorDesc(err))
			}
		} else if err == nil {
			return grpc.Errorf(codes.Internal, "nil error on empty read: io.ReaderAt contract violated")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return grpc.Errorf(codes.Unknown, "ReadAt(resourceName=%q): %v", req.ResourceName, err)
		}
	}

	log.Printf("Done streaming %q", req.ResourceName)

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Printf("Listen(): %v", err)
		return
	}

	g := grpc.NewServer()
	pb.RegisterByteStreamServer(g, bytestreamSrv{})

	log.Println("Serving on localhost:10000")
	log.Fatalln(g.Serve(lis))
}
