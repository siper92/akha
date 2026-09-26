package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/siper92/akha/platform/backend"
	auth2 "github.com/siper92/akha/platform/backend/auth"
	"github.com/siper92/akha/platform/backend/server"
	"github.com/siper92/akha/platform/sdk/db-sdk"
	"github.com/siper92/akha/platform/worker"
	"github.com/spf13/cobra"

	"github.com/siper92/akha/internal/config"
	"github.com/siper92/akha/lang/runner"
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
	db, err := auth2.OpenDB(ctx, cfg.DBPath)
	if err != nil {
		return err
	}

	defer db.Close()

	//@TODO: seed must be a SQL on just level
	store := auth2.NewStore(db_sdk.New(db))
	if err := store.Seed(ctx, cfg.AccessTokens); err != nil {
		return err
	}

	kp, err := auth2.LoadOrCreateKeys(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath)
	if err != nil {
		return err
	}

	keys := auth2.NewKeyring(kp, cfg.JWT.TTL)
	authn := auth2.NewAuthenticator(keys, store, log)

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
