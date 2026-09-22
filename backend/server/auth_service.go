package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/siper92/akha/backend/auth"
	proto_sdk "github.com/siper92/akha/sdk/proto-sdk"
)

type authService struct {
	proto_sdk.UnimplementedAuthServiceServer
	authn auth.Authenticator
}

var _ proto_sdk.AuthServiceServer = (*authService)(nil)

func newAuthService(authn auth.Authenticator) proto_sdk.AuthServiceServer {
	return &authService{authn: authn}
}

func (s *authService) Login(ctx context.Context, req *proto_sdk.LoginRequest) (*proto_sdk.LoginResponse, error) {
	token, _, err := s.authn.Register(ctx, req.GetAccessToken())
	if errors.Is(err, auth.ErrUnauthorized) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto_sdk.LoginResponse{Token: token}, nil
}

func (s *authService) ValidateToken(ctx context.Context, req *proto_sdk.TokenMessage) (*proto_sdk.ValidateTokenResponse, error) {
	_, err := s.authn.Verify(ctx, req.GetToken())
	return &proto_sdk.ValidateTokenResponse{Valid: err == nil}, nil
}
