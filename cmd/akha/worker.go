package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	worker2 "github.com/siper92/akha/platform/worker"
	client2 "github.com/siper92/akha/platform/worker/client"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/siper92/akha/internal/cache"
	"github.com/siper92/akha/internal/config"
	"github.com/siper92/akha/lang/check"
	"github.com/siper92/akha/lang/runner"
)

const backendTimeout = 10 * time.Second

func newWorkerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "worker commands",
	}
	configFlag(cmd, "config.wk.yaml")

	cmd.AddCommand(
		&cobra.Command{
			Use:   "run FILE",
			Short: "log in to the backend and run a .ak script",
			Args:  cobra.ExactArgs(1),
			RunE:  runWorkerRun,
		},
		&cobra.Command{
			Use:   "check FILE",
			Short: "static check a .ak script",
			Args:  cobra.ExactArgs(1),
			RunE:  runWorkerCheck,
		},
		&cobra.Command{
			Use:   "login",
			Short: "log in to the backend and cache the token",
			RunE:  runWorkerLogin,
		},
		&cobra.Command{
			Use:   "whoami",
			Short: "validate the cached token against the backend",
			RunE:  runWorkerWhoami,
		},
	)

	return cmd
}

type workerParts struct {
	cfg  config.Worker
	conn *grpc.ClientConn
	be   client2.Backend
	ts   client2.TokenSource
	w    worker2.Worker
}

func newWorkerParts(cmd *cobra.Command, log *slog.Logger) (*workerParts, error) {
	cfg, err := config.LoadWorker(configPath(cmd))
	if err != nil {
		return nil, err
	}

	c, err := cache.NewFile(cfg.CacheDir)
	if err != nil {
		return nil, err
	}

	conn, err := client2.Dial(cfg.Backend)
	if err != nil {
		return nil, err
	}

	be := client2.New(conn, cfg.AccessToken)
	ts := client2.NewTokenSource(be, c, client2.DefaultTokenKey)

	return &workerParts{
		cfg:  cfg,
		conn: conn,
		be:   be,
		ts:   ts,
		w:    worker2.New(ts, be, runner.New(log), log),
	}, nil
}

func (p *workerParts) close() {
	_ = p.conn.Close()
}

func runWorkerRun(cmd *cobra.Command, args []string) error {
	log := slog.Default()
	src, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	p, err := newWorkerParts(cmd, log)
	if err != nil {
		return err
	}
	defer p.close()
	ctx, cancel := context.WithTimeout(cmd.Context(), backendTimeout)
	if err := p.w.Start(ctx); err != nil {
		cancel()
		return err
	}
	cancel()

	res, err := p.w.Run(cmd.Context(), string(src), runner.Options{
		Root:     p.cfg.Root,
		CacheDir: p.cfg.CacheDir,
	})
	printDiagnostics(args[0], res.Diagnostics)
	if res.LogPath != "" {
		fmt.Fprintln(cmd.OutOrStdout(), "log:", res.LogPath)
		fmt.Fprintln(cmd.OutOrStdout(), "debug:", res.DebugPath)
	}
	switch {
	case errors.Is(err, runner.ErrParse), errors.Is(err, runner.ErrCheck):
		return &exitError{code: res.ExitCode}
	case err != nil:
		return &exitError{code: res.ExitCode, err: err}
	case res.ExitCode != 0:
		return &exitError{code: res.ExitCode}
	}
	fmt.Fprintln(cmd.OutOrStdout(), "exit:", res.ExitCode)
	return nil
}

func runWorkerCheck(cmd *cobra.Command, args []string) error {
	src, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	diags, err := runner.New(slog.Default()).Check(cmd.Context(), string(src))
	if err != nil {
		return err
	}
	printDiagnostics(args[0], diags)
	if len(diags) > 0 {
		return &exitError{code: runner.ExitCheck}
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok:", args[0])
	return nil
}

func runWorkerLogin(cmd *cobra.Command, args []string) error {
	p, err := newWorkerParts(cmd, slog.Default())
	if err != nil {
		return err
	}
	defer p.close()
	ctx, cancel := context.WithTimeout(cmd.Context(), backendTimeout)
	defer cancel()
	if err := p.ts.Invalidate(ctx); err != nil {
		return err
	}
	if _, err := p.ts.Token(ctx); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "login ok:", p.cfg.Backend)
	return nil
}

func runWorkerWhoami(cmd *cobra.Command, args []string) error {
	p, err := newWorkerParts(cmd, slog.Default())
	if err != nil {
		return err
	}
	defer p.close()
	ctx, cancel := context.WithTimeout(cmd.Context(), backendTimeout)
	defer cancel()
	token, err := p.ts.Token(ctx)
	if err != nil {
		return err
	}
	ok, err := p.be.Validate(ctx, token)
	if err != nil {
		return err
	}
	if !ok {
		return &exitError{code: 1, err: worker2.ErrAuth}
	}
	fmt.Fprintln(cmd.OutOrStdout(), "token valid:", p.cfg.Backend)
	return nil
}

func printDiagnostics(file string, diags []check.Diagnostic) {
	for _, d := range diags {
		sev := "error"
		if d.Severity == check.SeverityWarning {
			sev = "warning"
		}
		fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n", file, d.Pos.Line, d.Pos.Col, sev, d.Msg)
	}
}
