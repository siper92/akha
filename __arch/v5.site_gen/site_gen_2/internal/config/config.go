package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var ErrInvalidConfig = errors.New("invalid config")

type Config struct {
	MDDir      string `mapstructure:"md_dir"`
	Template   string `mapstructure:"template"`
	Texts      string `mapstructure:"texts"`
	StaticDir  string `mapstructure:"static_dir"`
	OutDir     string `mapstructure:"out_dir"`
	AssetsPath string `mapstructure:"assets_path"`
	IndexPage  string `mapstructure:"index_page"`
	BaseURL    string `mapstructure:"base_url"`
	Lang       string `mapstructure:"lang"`
	Clean      bool   `mapstructure:"clean"`
}

func Load(v *viper.Viper) (Config, error) {
	v.SetDefault("md_dir", "web/md")
	v.SetDefault("template", "web/templates/layout.html")
	v.SetDefault("texts", "web/templates/texts.yaml")
	v.SetDefault("static_dir", "web/static")
	v.SetDefault("out_dir", "deploy/static")
	v.SetDefault("assets_path", "/assets")
	v.SetDefault("index_page", "world")
	v.SetDefault("base_url", "")
	v.SetDefault("lang", "en")
	v.SetDefault("clean", true)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	cfg.AssetsPath = "/" + strings.Trim(cfg.AssetsPath, "/")

	fields := []struct{ name, val string }{
		{"md_dir", cfg.MDDir},
		{"template", cfg.Template},
		{"texts", cfg.Texts},
		{"static_dir", cfg.StaticDir},
		{"out_dir", cfg.OutDir},
		{"index_page", cfg.IndexPage},
		{"lang", cfg.Lang},
	}
	for _, f := range fields {
		if f.val == "" {
			return Config{}, fmt.Errorf("%w: %s is empty", ErrInvalidConfig, f.name)
		}
	}
	if cfg.AssetsPath == "/" {
		return Config{}, fmt.Errorf("%w: assets_path must not be the root", ErrInvalidConfig)
	}
	if cfg.BaseURL != "" && !strings.HasPrefix(cfg.BaseURL, "http://") && !strings.HasPrefix(cfg.BaseURL, "https://") {
		return Config{}, fmt.Errorf("%w: base_url must start with http:// or https://", ErrInvalidConfig)
	}
	return cfg, nil
}
