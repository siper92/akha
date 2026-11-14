package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"sitegen/internal/config"
	"sitegen/internal/pages"
	"sitegen/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "render markdown pages and serve them",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().String("addr", ":8080", "listen address")
	serveCmd.Flags().String("md-dir", "md", "markdown directory")
	_ = viper.BindPFlag("addr", serveCmd.Flags().Lookup("addr"))
	_ = viper.BindPFlag("md_dir", serveCmd.Flags().Lookup("md-dir"))
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, _ []string) error {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(log)

	cfg, err := config.Load(viper.GetViper())
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	src := pages.NewDirSource(cfg.MDDir, cfg.IndexPage)
	docs, err := src.List(ctx)
	if err != nil {
		return fmt.Errorf("list docs: %w", err)
	}

	rend, err := pages.NewRenderer(cfg.Template, cfg.SiteName, pages.NavFrom(docs))
	if err != nil {
		return err
	}

	store, err := pages.Build(ctx, src, rend)
	if err != nil {
		return err
	}
	log.Info("pages loaded", "count", len(store.Pages()))

	engine := server.New(cfg, store)
	log.Info("listening", "addr", cfg.Addr)
	return server.Run(ctx, engine, cfg.Addr)
}
