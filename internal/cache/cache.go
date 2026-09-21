package cache

import "context"

type Cache interface {
	Dir() string
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Put(ctx context.Context, key string, val []byte) error
	Del(ctx context.Context, key string) error
}
