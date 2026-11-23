package main

import (
	"context"
	"errors"
	"os"

	"aiflow/ai"
	"aiflow/config"
	"aiflow/debug"
	"aiflow/flow"
	"aiflow/tools"
	"aiflow/vfs"
)

var errTaskFailed = errors.New("task failed")

var tasks = []flow.TaskRecord{
	{
		ID:          "1",
		Title:       "create a short story",
		Description: "write a short story of about 300 words and save it in story.txt",
	},
	{
		ID:          "2",
		Title:       "split the story",
		Description: "read story.txt, split it in 2 parts of roughly equal length, save them in part1.txt and part2.txt, then delete story.txt",
	},
	{
		ID:          "3",
		Title:       "summarize the story",
		Description: "read part1.txt and part2.txt, write a summary of the story and save it in summary.txt",
	},
}

func main() {
	if err := run(context.Background()); err != nil {
		debug.Service().Log("error: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	files, err := vfs.NewDisk(config.GetWorkDir())
	if err != nil {
		return err
	}

	sp := flow.NewServices(
		flow.NewMemoryRepository(tasks...),
		flow.NewTaskServices(ai.NewAnthropic(), tools.Files(files)),
	)

	tasker, err := flow.NewTasker(sp)
	if err != nil {
		return err
	}

	state := &flow.State{}
	if err := tasker.Run(ctx, state); err != nil {
		return err
	}

	debug.Service().Log("processed %d of %d tasks", len(state.Processed), len(tasks))
	if state.Current.Status == flow.StatusFailed {
		return errTaskFailed
	}
	return nil
}
