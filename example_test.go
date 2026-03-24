package grpcpool_test

import (
	"fmt"

	"github.com/go-coldbrew/grpcpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ExampleNew() {
	// Create individual gRPC connections (NewClient is lazy — no real connection yet)
	conn1, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	conn2, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	// Create a round-robin pool from existing connections
	pool := grpcpool.New([]*grpc.ClientConn{conn1, conn2})
	defer pool.Close()

	fmt.Println("pool size:", pool.Num())
}
