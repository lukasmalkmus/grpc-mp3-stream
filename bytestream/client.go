package bytestream

import (
	"context"
	"math/rand"
	"time"

	"google.golang.org/grpc"

	pb "github.com/lukasmalkmus/stream/proto"
)

const (
	// MaxBufSize is the maximum buffer size (in bytes) received in a read
	// chunk.
	MaxBufSize  = 1024 * 64 // 64KiB
	backoffBase = time.Millisecond * 10
	backoffMax  = time.Second * 1
	maxTries    = 5
)

// Client is the go wrapper around a ByteStreamClient and provides an interface
// to it.
type Client struct {
	client  pb.ByteStreamClient
	options []grpc.CallOption
}

// NewClient creates a new Client.
func NewClient(cc *grpc.ClientConn, options ...grpc.CallOption) *Client {
	return &Client{
		client:  pb.NewByteStreamClient(cc),
		options: options,
	}
}

// Reader reads from a byte stream.
type Reader struct {
	ctx          context.Context
	c            *Client
	readClient   pb.ByteStream_ReadClient
	resourceName string
	err          error
	buf          []byte
}

// NewReader creates a new Reader to read a resource.
func (c *Client) NewReader(ctx context.Context, resourceName string) (*Reader, error) {
	// readClient is set up for Read(). ReadAt() will copy needed fields into
	// its reentrantReader.
	readClient, err := c.client.Read(ctx, &pb.ReadRequest{
		ResourceName: resourceName,
	}, c.options...)
	if err != nil {
		return nil, err
	}

	return &Reader{
		ctx:          ctx,
		c:            c,
		resourceName: resourceName,
		readClient:   readClient,
	}, nil
}

// ResourceName gets the resource name this Reader is reading.
func (r *Reader) ResourceName() string {
	return r.resourceName
}

// Read implements io.Reader.
// Read buffers received bytes that do not fit in p.
func (r *Reader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	var backoffDelay time.Duration
	for tries := 0; len(r.buf) == 0 && tries < maxTries; tries++ {
		// No data in buffer.
		resp, err := r.readClient.Recv()
		if err != nil {
			r.err = err
			return 0, err
		}
		r.buf = resp.Data
		if len(r.buf) != 0 {
			break
		}

		// back off
		if backoffDelay < backoffBase {
			backoffDelay = backoffBase
		} else {
			backoffDelay = time.Duration(float64(backoffDelay) * 1.3 * (1 - 0.4*rand.Float64()))
		}
		if backoffDelay > backoffMax {
			backoffDelay = backoffMax
		}
		select {
		case <-time.After(backoffDelay):
		case <-r.ctx.Done():
			if err := r.ctx.Err(); err != nil {
				r.err = err
			}
			return 0, r.err
		}
	}

	// Copy from buffer.
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// Close implements io.Closer.
func (r *Reader) Close() error {
	if r.readClient == nil {
		return nil
	}
	err := r.readClient.CloseSend()
	r.readClient = nil
	return err
}
