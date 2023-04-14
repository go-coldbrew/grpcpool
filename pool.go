// grpcpool is a pool of grpc.ClientConns. It implements grpc.ClientConnInterface to enable it to be used directly with generated proto stubs. It is based on https://github.com/googleapis/google-api-go-client/blob/v0.115.0/transport/grpc/pool.go
package grpcpool

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
)

// based on https://github.com/googleapis/google-api-go-client/blob/v0.115.0/transport/grpc/pool.go

// ConnPool is a pool of grpc.ClientConns.
type ConnPool interface {
	// Conn returns a ClientConn from the pool.
	//
	// Conns aren't returned to the pool.
	Conn() *grpc.ClientConn

	// Num returns the number of connections in the pool.
	//
	// It will always return the same value.
	Num() int

	// Close closes every ClientConn in the pool.
	//
	// The error returned by Close may be a single error or multiple errors.
	Close() error

	// ConnPool implements grpc.ClientConnInterface to enable it to be used directly with generated proto stubs.
	grpc.ClientConnInterface
}

var _ ConnPool = &roundRobinConnPool{}

type roundRobinConnPool struct {
	conns []*grpc.ClientConn

	idx uint32 // access via sync/atomic
}

func (p *roundRobinConnPool) Num() int {
	return len(p.conns)
}

func (p *roundRobinConnPool) Conn() *grpc.ClientConn {
	i := atomic.AddUint32(&p.idx, 1)
	return p.conns[i%uint32(len(p.conns))]
}

func (p *roundRobinConnPool) Close() error {
	var errs error
	for _, conn := range p.conns {
		if err := conn.Close(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (p *roundRobinConnPool) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return p.Conn().Invoke(ctx, method, args, reply, opts...)
}

func (p *roundRobinConnPool) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return p.Conn().NewStream(ctx, desc, method, opts...)
}

// New creates a new ConnPool from the given connections.
func New(conns []*grpc.ClientConn) ConnPool {
	if len(conns) == 0 {
		return nil
	}
	return &roundRobinConnPool{conns: conns}
}

// DialContext creates a new ConnPool with num connections to target.
func DialContext(ctx context.Context, target string, num uint, opts ...grpc.DialOption) (ConnPool, error) {
	if num == 0 {
		return nil, errors.New("grpcpool: num must be greater than 0")
	}
	conns := make([]*grpc.ClientConn, num)
	for i := range conns {
		conn, err := grpc.DialContext(ctx, target, opts...)
		if err != nil {
			return nil, err
		}
		conns[i] = conn
	}
	return New(conns), nil
}

// Dial creates a new ConnPool with num connections to target.
func Dial(target string, num uint, opts ...grpc.DialOption) (ConnPool, error) {
	return DialContext(context.Background(), target, num, opts...)
}
