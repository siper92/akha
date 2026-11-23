package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"sitegen/internal/config"
	"sitegen/internal/pages"
)

var cfgFile string

func main() {
	root := &cobra.Command{
		Use:           "site",
		Short:         "build the static html site from web/ into the output dir",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          run,
	}
	root.Flags().StringVar(&cfgFile, "config", "", "config file (default config.yaml in working dir)")
	root.Flags().String("out", "deploy/static", "output dir")
	root.Flags().String("base-url", "", "public site url, enables canonical, og:url and sitemap.xml")
	_ = viper.BindPFlag("out_dir", root.Flags().Lookup("out"))
	_ = viper.BindPFlag("base_url", root.Flags().Lookup("base-url"))

	if err := root.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	viper.SetEnvPrefix("SITE")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil && cfgFile != "" {
		return fmt.Errorf("config: %w", err)
	}

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(log)

	if err := readConfig(); err != nil {
		return err
	}
	cfg, err := config.Load(viper.GetViper())
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	docs, err := pages.NewDirSource(cfg.MDDir, cfg.IndexPage).List(ctx)
	if err != nil {
		return fmt.Errorf("list docs: %w", err)
	}

	texts, err := pages.LoadTexts(ctx, cfg.Texts)
	if err != nil {
		return err
	}

	rend, err := pages.NewRenderer(pages.RenderOptions{
		Template: cfg.Template,
		Texts:    texts,
		BaseURL:  cfg.BaseURL,
		Assets:   cfg.AssetsPath,
		Lang:     cfg.Lang,
		Nav:      pages.NavFrom(docs, texts.Menu),
	})
	if err != nil {
		return err
	}

	w := pages.NewDirWriter(cfg.OutDir, cfg.BaseURL)
	if cfg.Clean {
		if err := w.Reset(ctx); err != nil {
			return err
		}
	}
	assets := pages.NewDirAssets(cfg.StaticDir, filepath.Join(cfg.OutDir, filepath.FromSlash(strings.TrimPrefix(cfg.AssetsPath, "/"))))

	built, err := pages.Build(ctx, docs, rend, w, assets)
	if err != nil {
		return err
	}
	for _, p := range built {
		log.Info("page", "path", p.Path, "file", p.File, "title", p.Meta.Title)
	}
	log.Info("site built", "pages", len(built), "out", cfg.OutDir)

	return nil
}
