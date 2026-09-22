package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultBackendAddr = "localhost:50051"
	DefaultListenAddr  = ":50051"
	DefaultCacheDir    = ".cache/akha"
	DefaultRoot        = "."
	DefaultDBPath      = ".cache/akha/backend.db"
	DefaultPrivateKey  = ".cache/akha/jwt.key"
	DefaultPublicKey   = ".cache/akha/jwt.pub"
	DefaultTTL         = 15 * time.Minute
)

var (
	ErrRead   = errors.New("read config")
	ErrDecode = errors.New("decode config")
)

type defaulter interface {
	setDefaults()
}

type loader[T any] struct{}

var (
	_ Loader[Worker]  = (*loader[Worker])(nil)
	_ Loader[Backend] = (*loader[Backend])(nil)
	_ defaulter       = (*Worker)(nil)
	_ defaulter       = (*Backend)(nil)
)

func NewLoader[T any]() Loader[T] {
	return &loader[T]{}
}

func (l *loader[T]) Load(path string) (T, error) {
	var out T
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return out, fmt.Errorf("%w: %s: %v", ErrRead, path, err)
	}
	if err := v.Unmarshal(&out); err != nil {
		return out, fmt.Errorf("%w: %s: %v", ErrDecode, path, err)
	}
	if d, ok := any(&out).(defaulter); ok {
		d.setDefaults()
	}
	return out, nil
}

func LoadWorker(path string) (Worker, error) {
	return NewLoader[Worker]().Load(path)
}

func LoadBackend(path string) (Backend, error) {
	return NewLoader[Backend]().Load(path)
}

func (w *Worker) setDefaults() {
	if w.Backend == "" {
		w.Backend = DefaultBackendAddr
	}
	if w.CacheDir == "" {
		w.CacheDir = DefaultCacheDir
	}
	if w.Root == "" {
		w.Root = DefaultRoot
	}
}

func (b *Backend) setDefaults() {
	if b.Addr == "" {
		b.Addr = DefaultListenAddr
	}
	if b.DBPath == "" {
		b.DBPath = DefaultDBPath
	}
	if b.CacheDir == "" {
		b.CacheDir = DefaultCacheDir
	}
	if b.JWT.PrivateKeyPath == "" {
		b.JWT.PrivateKeyPath = DefaultPrivateKey
	}
	if b.JWT.PublicKeyPath == "" {
		b.JWT.PublicKeyPath = DefaultPublicKey
	}
	if b.JWT.TTL <= 0 {
		b.JWT.TTL = DefaultTTL
	}
}
