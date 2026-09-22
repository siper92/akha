package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const entryExt = ".json"

type fileCache struct {
	dir string
}

var _ Cache = (*fileCache)(nil)

func NewFile(dir string) (Cache, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &fileCache{dir: dir}, nil
}

func (c *fileCache) Dir() string { return c.dir }

func (c *fileCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	e, ok, err := c.Entry(ctx, key)
	if err != nil || !ok {
		return nil, false, err
	}
	return e.Value, true, nil
}

func (c *fileCache) Entry(ctx context.Context, key string) (Entry, bool, error) {
	e, ok, err := c.read(c.path(key))
	if err != nil || !ok {
		return Entry{}, false, err
	}
	if expired(e, time.Now()) {
		return Entry{}, false, c.Del(ctx, key)
	}
	return e, true, nil
}

func (c *fileCache) Put(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	e := Entry{Key: key, Value: val}
	if ttl != NoExpiry {
		e.ExpiresAt = time.Now().Add(ttl)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	tmp := c.path(key) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.path(key))
}

func (c *fileCache) Del(ctx context.Context, key string) error {
	err := os.Remove(c.path(key))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (c *fileCache) Purge(ctx context.Context) (int, error) {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	n := 0
	for _, d := range entries {
		if d.IsDir() || !strings.HasSuffix(d.Name(), entryExt) {
			continue
		}
		p := filepath.Join(c.dir, d.Name())
		e, ok, err := c.read(p)
		if err != nil || !ok || !expired(e, now) {
			continue
		}
		if err := os.Remove(p); err == nil {
			n++
		}
	}
	return n, nil
}

func (c *fileCache) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(c.dir, hex.EncodeToString(sum[:16])+entryExt)
}

func (c *fileCache) read(p string) (Entry, bool, error) {
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, err
	}
	var e Entry
	if err := json.Unmarshal(b, &e); err != nil {
		return Entry{}, false, err
	}
	return e, true, nil
}

func expired(e Entry, now time.Time) bool {
	return !e.ExpiresAt.IsZero() && now.After(e.ExpiresAt)
}
