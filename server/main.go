package main

import (
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lukasmalkmus/stream/bytestream"
	pb "github.com/lukasmalkmus/stream/proto"
)

type bytestreamSrv struct{}

// Read is used to retrieve the contents of a resource as a sequence of bytes.
// The bytes are returned in a sequence of responses, and the responses are
// delivered as the results of a server-side streaming RPC.
func (b *bytestreamSrv) Read(req *pb.ReadRequest, stream pb.ByteStream_ReadServer) error {
	if req.ResourceName == "" {
		return status.Errorf(codes.InvalidArgument, "ReadRequest: empty or missing resource_name")
	}

	f, err := os.Open(req.ResourceName)
	if err != nil {
		return status.Errorf(codes.NotFound, "ReadRequest: resource %q doesn't exist", req.ResourceName)
	}
	defer f.Close()

	buf := make([]byte, bytestream.MaxBufSize)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if err := stream.Send(&pb.ReadResponse{Data: buf[:n]}); err != nil {
				return status.Errorf(status.Code(err), "Send(resourceName=%q): %v", req.ResourceName, status.Convert(err).Message())
			}
		} else if err == nil {
			return status.Errorf(codes.Internal, "nil error on empty read: io.ReaderAt contract violated")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "ReadAt(resourceName=%q): %v", req.ResourceName, err)
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
	pb.RegisterByteStreamServer(g, &bytestreamSrv{})

	log.Println("Serving on", lis.Addr().String())
	log.Fatalln(g.Serve(lis))
}
