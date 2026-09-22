package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/siper92/akha/internal/tu"
)

func writeYAML(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadWorker(t *testing.T) {
	full := writeYAML(t, "config.wk.yaml", `
backend: "127.0.0.1:1234"
worker_access_token: "tok"
cache_dir: "/tmp/c"
root: "/tmp/r"
`)
	minimal := writeYAML(t, "min.yaml", `worker_access_token: "tok"`)

	// --- explicit values and defaults
	cases := []tu.Case[string, Worker]{
		{
			Name:     "all_fields",
			Input:    full,
			Expected: Worker{Backend: "127.0.0.1:1234", AccessToken: "tok", CacheDir: "/tmp/c", Root: "/tmp/r"},
		},
		{
			Name:     "defaults_applied",
			Input:    minimal,
			Expected: Worker{Backend: DefaultBackendAddr, AccessToken: "tok", CacheDir: DefaultCacheDir, Root: DefaultRoot},
		},
		{
			Name:  "missing_file",
			Input: filepath.Join(t.TempDir(), "nope.yaml"),
			Err:   ErrRead,
		},
	}
	tu.Run(tu.New(t), cases, LoadWorker, nil)
}

func TestLoadBackend(t *testing.T) {
	full := writeYAML(t, "config.be.yaml", `
addr: ":6000"
db: "/tmp/db.sqlite"
cache_dir: "/tmp/c"
workers: 2
access_tokens:
  - "a"
  - "b"
jwt:
  private_key: "/tmp/k"
  public_key: "/tmp/p"
  ttl: "1h"
`)
	minimal := writeYAML(t, "min.yaml", `access_tokens: ["a"]`)

	// --- explicit values and defaults
	cases := []tu.Case[string, Backend]{
		{
			Name:  "all_fields",
			Input: full,
			Expected: Backend{
				Addr: ":6000", DBPath: "/tmp/db.sqlite", CacheDir: "/tmp/c", Workers: 2,
				AccessTokens: []string{"a", "b"},
				JWT:          JWT{PrivateKeyPath: "/tmp/k", PublicKeyPath: "/tmp/p", TTL: time.Hour},
			},
		},
		{
			Name:  "defaults_applied",
			Input: minimal,
			Expected: Backend{
				Addr: DefaultListenAddr, DBPath: DefaultDBPath, CacheDir: DefaultCacheDir,
				AccessTokens: []string{"a"},
				JWT:          JWT{PrivateKeyPath: DefaultPrivateKey, PublicKeyPath: DefaultPublicKey, TTL: DefaultTTL},
			},
		},
		{
			Name:  "missing_file",
			Input: filepath.Join(t.TempDir(), "nope.yaml"),
			Err:   ErrRead,
		},
	}
	tu.Run(tu.New(t), cases, LoadBackend, nil)
}
