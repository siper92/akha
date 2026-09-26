package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	auth2 "github.com/siper92/akha/platform/backend/auth"
	"github.com/siper92/akha/platform/sdk/db-sdk"
	proto_sdk2 "github.com/siper92/akha/platform/sdk/proto-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/siper92/akha/internal/tu"
)

const accessToken = "akha_test_token"

func startServer(t *testing.T) proto_sdk2.AuthServiceClient {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	db, err := auth2.OpenDB(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	st := auth2.NewStore(db_sdk.New(db))
	if err := st.Seed(ctx, []string{accessToken}); err != nil {
		t.Fatal(err)
	}

	kp, err := auth2.GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	authn := auth2.NewAuthenticator(auth2.NewKeyring(kp, time.Minute), st, log)

	lis := bufconn.Listen(1 << 20)
	srv := New(lis, authn, log)
	go func() { _ = srv.Serve(ctx) }()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	return proto_sdk2.NewAuthServiceClient(conn)
}

func TestLogin(t *testing.T) {
	cli := startServer(t)
	ctx := context.Background()

	// --- login outcomes as grpc codes
	cases := []tu.Case[string, codes.Code]{
		{
			Name:     "known_access_token",
			Input:    accessToken,
			Expected: codes.OK,
		},
		{
			Name:     "unknown_access_token",
			Input:    "nope",
			Expected: codes.Unauthenticated,
		},
	}

	fn := func(tok string) (codes.Code, error) {
		resp, err := cli.Login(ctx, &proto_sdk2.LoginRequest{AccessToken: tok})
		if err != nil {
			return status.Code(err), nil
		}
		if resp.GetToken() == "" {
			return codes.Unknown, nil
		}
		return codes.OK, nil
	}
	tu.Run(tu.New(t), cases, fn, nil)
}

func TestValidateToken(t *testing.T) {
	cli := startServer(t)
	ctx := context.Background()
	resp, err := cli.Login(ctx, &proto_sdk2.LoginRequest{AccessToken: accessToken})
	if err != nil {
		t.Fatal(err)
	}

	// --- validation results
	cases := []tu.Case[string, bool]{
		{
			Name:     "issued_token",
			Input:    resp.GetToken(),
			Expected: true,
		},
		{
			Name:     "garbage_token",
			Input:    "garbage",
			Expected: false,
		},
		{
			Name:     "empty_token",
			Input:    "",
			Expected: false,
		},
	}
	fn := func(tok string) (bool, error) {
		r, err := cli.ValidateToken(ctx, &proto_sdk2.TokenMessage{Token: tok})
		if err != nil {
			return false, err
		}
		return r.GetValid(), nil
	}
	tu.Run(tu.New(t), cases, fn, nil)
}

func TestBearer(t *testing.T) {
	// --- header parsing
	cases := []tu.Case[string, string]{
		{
			Name:     "bearer_lower",
			Input:    "bearer abc",
			Expected: "abc",
		},
		{
			Name:     "bearer_upper",
			Input:    "Bearer abc",
			Expected: "abc",
		},
		{
			Name:     "no_prefix",
			Input:    "abc",
			Expected: "",
		},
		{
			Name:     "empty",
			Input:    "",
			Expected: "",
		},
	}
	fn := func(h string) (string, error) {
		ctx := context.Background()
		if h != "" {
			ctx = withIncomingAuth(ctx, h)
		}
		return bearer(ctx), nil
	}
	tu.Run(tu.New(t), cases, fn, nil)
}
