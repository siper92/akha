package config

import "time"

type Worker struct {
	Backend     string `mapstructure:"backend"`
	AccessToken string `mapstructure:"worker_access_token"`
	CacheDir    string `mapstructure:"cache_dir"`
	Root        string `mapstructure:"root"`
}

type JWT struct {
	PrivateKeyPath string        `mapstructure:"private_key"`
	PublicKeyPath  string        `mapstructure:"public_key"`
	TTL            time.Duration `mapstructure:"ttl"`
}

type Backend struct {
	Addr         string   `mapstructure:"addr"`
	DBPath       string   `mapstructure:"db"`
	CacheDir     string   `mapstructure:"cache_dir"`
	Workers      int      `mapstructure:"workers"`
	AccessTokens []string `mapstructure:"access_tokens"`
	JWT          JWT      `mapstructure:"jwt"`
}

type Loader[T any] interface {
	Load(path string) (T, error)
}
