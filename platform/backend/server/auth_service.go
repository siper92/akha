package server

import (
	"context"
	"errors"

	auth2 "github.com/siper92/akha/platform/backend/auth"
	proto_sdk2 "github.com/siper92/akha/platform/sdk/proto-sdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authService struct {
	proto_sdk2.UnimplementedAuthServiceServer
	authn auth2.Authenticator
}

var _ proto_sdk2.AuthServiceServer = (*authService)(nil)

func newAuthService(authn auth2.Authenticator) proto_sdk2.AuthServiceServer {
	return &authService{authn: authn}
}

func (s *authService) Login(ctx context.Context, req *proto_sdk2.LoginRequest) (*proto_sdk2.LoginResponse, error) {
	token, _, err := s.authn.Register(ctx, req.GetAccessToken())
	if errors.Is(err, auth2.ErrUnauthorized) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto_sdk2.LoginResponse{Token: token}, nil
}

func (s *authService) ValidateToken(ctx context.Context, req *proto_sdk2.TokenMessage) (*proto_sdk2.ValidateTokenResponse, error) {
	_, err := s.authn.Verify(ctx, req.GetToken())
	return &proto_sdk2.ValidateTokenResponse{Valid: err == nil}, nil
}
