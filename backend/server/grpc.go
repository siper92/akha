package server

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/siper92/akha/backend/auth"
	proto_sdk "github.com/siper92/akha/sdk/proto-sdk"
)

type server struct {
	lis net.Listener
	gs  *grpc.Server
	log *slog.Logger
}

var _ Server = (*server)(nil)

func New(lis net.Listener, authn auth.Authenticator, log *slog.Logger) Server {
	if log == nil {
		log = slog.Default()
	}
	authI := NewAuthInterceptor(authn,
		proto_sdk.AuthService_Login_FullMethodName,
		proto_sdk.AuthService_ValidateToken_FullMethodName,
	)
	logI := NewLogInterceptor(log)
	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logI.Unary(), authI.Unary()),
		grpc.ChainStreamInterceptor(logI.Stream(), authI.Stream()),
	)
	proto_sdk.RegisterAuthServiceServer(gs, newAuthService(authn))
	return &server{lis: lis, gs: gs, log: log}
}

func (s *server) Addr() string { return s.lis.Addr().String() }

func (s *server) Serve(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			s.gs.GracefulStop()
		case <-done:
		}
	}()
	s.log.Info("backend listening", "addr", s.Addr())
	err := s.gs.Serve(s.lis)
	close(done)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}
