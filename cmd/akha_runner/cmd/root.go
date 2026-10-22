package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

const (
	flagConfig  = "config"
	flagVerbose = "verbose"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "akha_runner",
		Short:         "akha runner - a AI workflow runner for task execution and planning",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool(flagVerbose)
			setupLogger(verbose)
		},
	}
	root.PersistentFlags().BoolP(flagVerbose, "v", false, "debug logging")
	root.AddCommand(
		newTaskCmd(),
	)

	return root
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
