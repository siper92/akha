package auth

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/siper92/akha/internal/tu"
	db_sdk "github.com/siper92/akha/sdk/db-sdk"
)

const goodToken = "akha_test_token"

type fixture struct {
	db    *sql.DB
	store Store
	keys  Keyring
	authn Authenticator
}

func newFixture(t *testing.T, ttl time.Duration) *fixture {
	t.Helper()

	ctx := context.Background()
	db, err := OpenDB(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { db.Close() })
	st := NewStore(db_sdk.New(db))

	if err := st.Seed(ctx, []string{goodToken}); err != nil {
		t.Fatal(err)
	}

	kp, err := GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}
	keys := NewKeyring(kp, ttl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &fixture{db: db, store: st, keys: keys, authn: NewAuthenticator(keys, st, log)}
}

func (f *fixture) count(t *testing.T, query string) int {
	t.Helper()
	var n int
	if err := f.db.QueryRow(query).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRegister(t *testing.T) {
	f := newFixture(t, time.Minute)
	ctx := WithAddr(context.Background(), "127.0.0.1:1")

	// --- access token lookup
	cases := []tu.Case[string, Tier]{
		{
			Name:     "known_token_issues_worker_jwt",
			Input:    goodToken,
			Expected: TierWorker,
		},
		{
			Name:  "unknown_token_is_rejected",
			Input: "nope",
			Err:   ErrUnauthorized,
		},
		{
			Name:  "empty_token_is_rejected",
			Input: "",
			Err:   ErrUnauthorized,
		},
	}

	fn := func(accessToken string) (Tier, error) {
		token, exp, err := f.authn.Register(ctx, accessToken)
		if err != nil {
			return "", err
		}

		if exp.Before(time.Now()) {
			t.Fatalf("expiry in the past: %v", exp)
		}
		c, err := f.authn.Verify(ctx, token)
		if err != nil {
			return "", err
		}
		if c.Subject != "worker-1" {
			t.Fatalf("subject: want worker-1, got %s", c.Subject)
		}
		return c.Tier, nil
	}
	tu.Run(tu.New(t), cases, fn, nil)

	if n := f.count(t, "SELECT count(*) FROM login_log"); n != 3 {
		t.Fatalf("login_log rows: want 3, got %d", n)
	}
	if n := f.count(t, "SELECT count(*) FROM login_log WHERE ok = 1"); n != 1 {
		t.Fatalf("successful logins: want 1, got %d", n)
	}
	if n := f.count(t, "SELECT count(*) FROM issued_tokens WHERE is_active = 1"); n != 1 {
		t.Fatalf("issued tokens: want 1, got %d", n)
	}
}

func TestVerify(t *testing.T) {
	f := newFixture(t, time.Minute)
	ctx := context.Background()

	token, _, err := f.authn.Register(ctx, goodToken)
	if err != nil {
		t.Fatal(err)
	}

	revoked, _, err := f.authn.Register(ctx, goodToken)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.store.Revoke(ctx, revoked); err != nil {
		t.Fatal(err)
	}

	keys, err := GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}

	foreign, _, err := NewKeyring(keys, time.Minute).Issue(ctx, "worker-1", TierWorker)
	if err != nil {
		t.Fatal(err)
	}

	expired, _, err := NewKeyring(keys, -time.Minute).Issue(ctx, "worker-1", TierWorker)
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(token, ".")
	tampered := parts[0] + "." + parts[1] + "x." + parts[2]

	// --- token states
	cases := []tu.Case[string, string]{
		{
			Name:     "issued_token_is_valid",
			Input:    token,
			Expected: "worker-1",
		},
		{
			Name:  "revoked_token",
			Input: revoked,
			Err:   ErrRevoked,
		},
		{
			Name:  "foreign_key_token",
			Input: foreign,
			Err:   ErrInvalidToken,
		},
		{
			Name:  "tampered_token",
			Input: tampered,
			Err:   ErrInvalidToken,
		},
		{
			Name:  "garbage_token",
			Input: "not.a.jwt",
			Err:   ErrInvalidToken,
		},
	}
	fn := func(tok string) (string, error) {
		c, err := f.authn.Verify(ctx, tok)
		return c.Subject, err
	}
	tu.Run(tu.New(t), cases, fn, nil)

	if _, err := NewKeyring(keys, time.Minute).Verify(ctx, expired); err != ErrExpiredToken {
		t.Fatalf("expired: want %v, got %v", ErrExpiredToken, err)
	}
}

func TestKeysRoundTrip(t *testing.T) {
	dir := t.TempDir()
	priv := filepath.Join(dir, "sub", "jwt.key")
	pub := filepath.Join(dir, "sub", "jwt.pub")

	first, err := LoadOrCreateKeys(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	second, err := LoadOrCreateKeys(priv, pub)
	if err != nil {
		t.Fatal(err)
	}

	if !first.Private.Equal(second.Private) || !first.Public.Equal(second.Public) {
		t.Fatal("keys must be loaded from disk on the second call")
	}

	ctx := context.Background()
	tok, _, err := NewKeyring(first, time.Minute).Issue(ctx, "s", TierBackend)

	if err != nil {
		t.Fatal(err)
	}
	c, err := NewKeyring(second, time.Minute).Verify(ctx, tok)
	if err != nil || c.Tier != TierBackend {
		t.Fatalf("verify with reloaded keys: %v %+v", err, c)
	}
}

func TestHashToken(t *testing.T) {
	// --- hashing is stable and distinct
	cases := []tu.Case[string, bool]{
		{
			Name:     "same_input_same_hash",
			Input:    "a",
			Expected: true,
		},
	}
	tu.Run(tu.New(t), cases, func(s string) (bool, error) {
		return HashToken(s) == HashToken(s) &&
			HashToken(s) != HashToken(s+"x") &&
			len(HashToken(s)) == 64, nil
	}, nil)
}
