package cmd

import "github.com/spf13/cobra"

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task FILE",
		Short: "task commands",
		Args:  cobra.ExactArgs(1),
		RunE:  runTaskRun,
	}
	configFlag(cmd, "config.task.yaml")

	return cmd
}

func runTaskRun(cmd *cobra.Command, args []string) error {

	return nil
}
