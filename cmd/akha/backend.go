package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/siper92/akha/backend"
	"github.com/siper92/akha/backend/auth"
	"github.com/siper92/akha/backend/server"
	"github.com/siper92/akha/internal/config"
	"github.com/siper92/akha/lang/runner"
	db_sdk "github.com/siper92/akha/sdk/db-sdk"
	"github.com/siper92/akha/worker"
)

func newBackendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend",
		Short: "backend commands",
	}
	configFlag(cmd, "config.be.yaml")
	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "start the backend gRPC server",
		RunE:  runBackendServe,
	})
	return cmd
}

func runBackendServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadBackend(configPath(cmd))
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log := slog.Default()

	db, err := auth.OpenDB(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()
	store := auth.NewStore(db_sdk.New(db))
	if err := store.Seed(ctx, cfg.AccessTokens); err != nil {
		return err
	}
	kp, err := auth.LoadOrCreateKeys(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath)
	if err != nil {
		return err
	}
	keys := auth.NewKeyring(kp, cfg.JWT.TTL)
	authn := auth.NewAuthenticator(keys, store, log)

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	srv := server.New(lis, authn, log)

	if cfg.Workers > 0 {
		sp := worker.NewSpawner(backend.EmbeddedFactory(keys, runner.New(log), log), log)
		if err := sp.Spawn(ctx, cfg.Workers); err != nil {
			return err
		}
	}
	log.Info("backend config", "db", cfg.DBPath, "workers", cfg.Workers, "ttl", cfg.JWT.TTL, "tokens", len(cfg.AccessTokens))
	return srv.Serve(ctx)
}

var _ context.Context = context.Background()
