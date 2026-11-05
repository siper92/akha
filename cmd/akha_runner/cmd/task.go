package cmd

import (
	"fmt"

	"github.com/siper92/akha/cmd/akha_runner/internal/task"
	"github.com/spf13/cobra"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task FILE",
		Short: "task commands - get task file and execute it in predefined steps",
		Args:  cobra.ExactArgs(1),
		RunE:  runTaskRun,
	}
	configFlag(cmd, "config.task.yaml")

	return cmd
}

func runTaskRun(cmd *cobra.Command, args []string) error {
	fmt.Println("Running task with file:", args[0])

	steps, err := task.GetTaskSteps()
	if err != nil {
		return err
	}

	flowSteps, err := task.ToFlowStep(steps)
	if err != nil {
		return err
	}

	for _, step := range flowSteps {
		cmd.Println("Step:", step.Name())
		cmd.Println()
	}

	return nil
}
