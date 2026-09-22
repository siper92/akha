package client

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	proto_sdk "github.com/siper92/akha/sdk/proto-sdk"
)

var ErrLogin = errors.New("login failed")

type backend struct {
	cli         proto_sdk.AuthServiceClient
	accessToken string
}

var _ Backend = (*backend)(nil)

func Dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func New(conn grpc.ClientConnInterface, accessToken string) Backend {
	return &backend{cli: proto_sdk.NewAuthServiceClient(conn), accessToken: accessToken}
}

func (b *backend) Login(ctx context.Context) (string, error) {
	resp, err := b.cli.Login(ctx, &proto_sdk.LoginRequest{AccessToken: b.accessToken})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLogin, err)
	}
	if resp.GetToken() == "" {
		return "", fmt.Errorf("%w: empty token", ErrLogin)
	}
	return resp.GetToken(), nil
}

func (b *backend) Validate(ctx context.Context, token string) (bool, error) {
	resp, err := b.cli.ValidateToken(ctx, &proto_sdk.TokenMessage{Token: token})
	if err != nil {
		return false, err
	}
	return resp.GetValid(), nil
}
