package cache

import (
	"context"
	"time"
)

const NoExpiry time.Duration = 0

type Entry struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
}

type Cache interface {
	Dir() string
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Entry(ctx context.Context, key string) (Entry, bool, error)
	Put(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Purge(ctx context.Context) (int, error)
}
