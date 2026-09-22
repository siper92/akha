package cache

import (
	"context"
	"testing"
	"time"

	"github.com/siper92/akha/internal/tu"
)

type putIn struct {
	key string
	val string
	ttl time.Duration
}

func TestFileCache(t *testing.T) {
	c, err := NewFile(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// --- put then get
	cases := []tu.Case[putIn, string]{
		{
			Name:     "no_expiry_is_kept",
			Input:    putIn{key: "a", val: "1", ttl: NoExpiry},
			Expected: "1",
		},
		{
			Name:     "future_ttl_is_kept",
			Input:    putIn{key: "b", val: "2", ttl: time.Hour},
			Expected: "2",
		},
		{
			Name:     "past_ttl_is_gone",
			Input:    putIn{key: "c", val: "3", ttl: -time.Second},
			Expected: "",
		},
		{
			Name:     "overwrite_replaces",
			Input:    putIn{key: "a", val: "9", ttl: NoExpiry},
			Expected: "9",
		},
	}

	fn := func(in putIn) (string, error) {
		if err := c.Put(ctx, in.key, []byte(in.val), in.ttl); err != nil {
			return "", err
		}
		b, ok, err := c.Get(ctx, in.key)
		if err != nil || !ok {
			return "", err
		}
		return string(b), nil
	}
	tu.Run(tu.New(t), cases, fn, nil)
}

func TestFileCacheDelAndPurge(t *testing.T) {
	c, err := NewFile(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if _, ok, _ := c.Get(ctx, "missing"); ok {
		t.Fatal("missing key must not be found")
	}

	if err := c.Del(ctx, "missing"); err != nil {
		t.Fatalf("del missing: %v", err)
	}
	if err := c.Put(ctx, "k", []byte("v"), NoExpiry); err != nil {
		t.Fatal(err)
	}
	if err := c.Del(ctx, "k"); err != nil {
		t.Fatal(err)
	}

	if _, ok, _ := c.Get(ctx, "k"); ok {
		t.Fatal("deleted key must not be found")
	}
	_ = c.Put(ctx, "live", []byte("1"), time.Hour)
	_ = c.Put(ctx, "dead1", []byte("1"), -time.Second)
	_ = c.Put(ctx, "dead2", []byte("1"), -time.Minute)
	n, err := c.Purge(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Fatalf("purge: want 2, got %d", n)
	}

	if _, ok, _ := c.Get(ctx, "live"); !ok {
		t.Fatal("live key must survive purge")
	}
}
