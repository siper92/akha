package server

import (
	"context"

	"google.golang.org/grpc"
)

type Server interface {
	Addr() string
	Serve(ctx context.Context) error
}

type Interceptor interface {
	Unary() grpc.UnaryServerInterceptor
	Stream() grpc.StreamServerInterceptor
}
