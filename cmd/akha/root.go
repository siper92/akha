package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

const (
	flagConfig  = "config"
	flagVerbose = "verbose"
)

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "akha",
		Short:         "akha - a platform for AI workflows",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool(flagVerbose)
			setupLogger(verbose)
		},
	}
	root.PersistentFlags().BoolP(flagVerbose, "v", false, "debug logging")
	root.AddCommand(newBackendCmd(), newWorkerCmd())
	return root
}

func newWorkerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "worker commands",
	}
	configFlag(cmd, "config.wk.yaml")
	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "start the worker gRPC server",
	})
	return cmd
}

func setupLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}

func configFlag(cmd *cobra.Command, def string) {
	cmd.PersistentFlags().String(flagConfig, def, "config file")
}

func configPath(cmd *cobra.Command) string {
	p, _ := cmd.Flags().GetString(flagConfig)
	return p
}
