package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/siper92/akha/platform/backend/auth"
	"github.com/siper92/akha/platform/backend/server"
	"github.com/siper92/akha/platform/sdk/db-sdk"
	"github.com/siper92/akha/platform/worker/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/siper92/akha/internal/cache"
	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/runner"
)

const (
	accessToken = "akha_test_token"
	helloScript = `
Ak.Allow(FS...)
Ak.Setup(log="hello.log", debug="hello.debug.log")
Ak.Log("hello from worker test")
FS.WriteFile("hello/hello.txt", "hello")
FS.UpdateFile("hello/hello.txt", "hello, world")
FS.ReadFile("hello/hello.txt")
FS.ListFiles("hello/")
Ak.Exit("done", code=0)
`
)

type env struct {
	conn  *grpc.ClientConn
	cache cache.Cache
	log   *slog.Logger
}

func startBackend(t *testing.T) *env {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	db, err := auth.OpenDB(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	st := auth.NewStore(db_sdk.New(db))
	if err := st.Seed(ctx, []string{accessToken}); err != nil {
		t.Fatal(err)
	}
	kp, err := auth.GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	authn := auth.NewAuthenticator(auth.NewKeyring(kp, time.Minute), st, log)
	lis := bufconn.Listen(1 << 20)
	go func() { _ = server.New(lis, authn, log).Serve(ctx) }()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	c, err := cache.NewFile(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return &env{conn: conn, cache: c, log: log}
}

func tokenKey(accessToken string) string {
	return client.DefaultTokenKey + "." + accessToken
}

func (e *env) worker(token string) (Worker, client.TokenSource) {
	be := client.New(e.conn, token)
	ts := client.NewTokenSource(be, e.cache, tokenKey(token))
	return New(ts, be, runner.New(e.log), e.log), ts
}

func TestStart(t *testing.T) {
	e := startBackend(t)

	// --- login with the static access token
	cases := []tu.Case[string, bool]{
		{
			Name:     "known_access_token",
			Input:    accessToken,
			Expected: true,
		},
		{
			Name:  "unknown_access_token",
			Input: "nope",
			Err:   ErrAuth,
		},
	}
	fn := func(tok string) (bool, error) {
		w, _ := e.worker(tok)
		if err := w.Start(context.Background()); err != nil {
			return false, err
		}
		return true, w.Stop(context.Background())
	}
	tu.Run(tu.New(t), cases, fn, nil)
}

func TestRunIntegration(t *testing.T) {
	e := startBackend(t)
	ctx := context.Background()
	w, ts := e.worker(accessToken)

	if _, err := w.Run(ctx, helloScript, runner.Options{}); !errors.Is(err, ErrNotStarted) {
		t.Fatalf("run before start: want %v, got %v", ErrNotStarted, err)
	}
	if err := w.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	first, err := ts.Token(ctx)
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	cacheDir := t.TempDir()
	res, err := w.Run(ctx, helloScript, runner.Options{RunID: "it", Root: root, CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code: want 0, got %d", res.ExitCode)
	}
	b, err := os.ReadFile(filepath.Join(root, "hello", "hello.txt"))
	if err != nil || string(b) != "hello, world" {
		t.Fatalf("hello.txt: %q %v", string(b), err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "runs", "it", "hello.log")); err != nil {
		t.Fatalf("log file: %v", err)
	}

	if err := w.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := w.Start(ctx); err != nil {
		t.Fatalf("restart: %v", err)
	}
	second, err := ts.Token(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("cached token must be reused across starts")
	}
}

func TestStartReplacesRejectedToken(t *testing.T) {
	e := startBackend(t)
	ctx := context.Background()
	if err := e.cache.Put(ctx, tokenKey(accessToken), []byte("stale"), cache.NoExpiry); err != nil {
		t.Fatal(err)
	}
	w, ts := e.worker(accessToken)
	if err := w.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	tok, err := ts.Token(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if tok == "stale" {
		t.Fatal("stale token must be replaced after rejection")
	}
}
