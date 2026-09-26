package server

import (
	"context"
	"log/slog"
	"strings"
	"time"

	auth2 "github.com/siper92/akha/platform/backend/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	authHeader   = "authorization"
	bearerPrefix = "bearer "
)

type claimsKey struct{}

func ClaimsFrom(ctx context.Context) (auth2.Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(auth2.Claims)
	return c, ok
}

type authInterceptor struct {
	verifier auth2.TokenVerifier
	open     map[string]bool
}

var _ Interceptor = (*authInterceptor)(nil)

func NewAuthInterceptor(v auth2.TokenVerifier, openMethods ...string) Interceptor {
	open := make(map[string]bool, len(openMethods))
	for _, m := range openMethods {
		open[m] = true
	}
	return &authInterceptor{verifier: v, open: open}
}

func (i *authInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, err := i.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *authInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := i.authorize(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
	}
}

func (i *authInterceptor) authorize(ctx context.Context, method string) (context.Context, error) {
	ctx = withPeerAddr(ctx)
	if i.open[method] {
		return ctx, nil
	}
	token := bearer(ctx)
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing bearer token")
	}
	c, err := i.verifier.Verify(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return context.WithValue(ctx, claimsKey{}, c), nil
}

func bearer(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, v := range md.Get(authHeader) {
		if strings.HasPrefix(strings.ToLower(v), bearerPrefix) {
			return strings.TrimSpace(v[len(bearerPrefix):])
		}
	}
	return ""
}

func withPeerAddr(ctx context.Context) context.Context {
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		return auth2.WithAddr(ctx, p.Addr.String())
	}
	return ctx
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }

type logInterceptor struct {
	log *slog.Logger
}

var _ Interceptor = (*logInterceptor)(nil)

func NewLogInterceptor(log *slog.Logger) Interceptor {
	if log == nil {
		log = slog.Default()
	}
	return &logInterceptor{log: log}
}

func (i *logInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		i.log.Info("rpc", "method", info.FullMethod, "code", status.Code(err).String(), "took", time.Since(start))
		return resp, err
	}
}

func (i *logInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)
		i.log.Info("rpc", "method", info.FullMethod, "code", status.Code(err).String(), "took", time.Since(start))
		return err
	}
}
