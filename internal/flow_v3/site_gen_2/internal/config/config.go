package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

var ErrInvalidConfig = errors.New("invalid config")

type Config struct {
	Addr      string `mapstructure:"addr"`
	MDDir     string `mapstructure:"md_dir"`
	Template  string `mapstructure:"template"`
	StaticDir string `mapstructure:"static_dir"`
	IndexPage string `mapstructure:"index_page"`
	SiteName  string `mapstructure:"site_name"`
}

func Load(v *viper.Viper) (Config, error) {
	v.SetDefault("addr", ":8080")
	v.SetDefault("md_dir", "md")
	v.SetDefault("template", "web/templates/layout.html")
	v.SetDefault("static_dir", "web/static")
	v.SetDefault("index_page", "world")
	v.SetDefault("site_name", "Akha Collective")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	fields := []struct{ name, val string }{
		{"addr", cfg.Addr},
		{"md_dir", cfg.MDDir},
		{"template", cfg.Template},
		{"static_dir", cfg.StaticDir},
		{"index_page", cfg.IndexPage},
		{"site_name", cfg.SiteName},
	}
	for _, f := range fields {
		if f.val == "" {
			return Config{}, fmt.Errorf("%w: %s is empty", ErrInvalidConfig, f.name)
		}
	}
	return cfg, nil
}
