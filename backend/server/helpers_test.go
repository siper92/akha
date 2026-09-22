package server

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func withIncomingAuth(ctx context.Context, value string) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.Pairs(authHeader, value))
}
