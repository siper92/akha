package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	flow "github.com/siper92/akha/projects/mcp/v3"
)

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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	debugFile, err := flow.OpenDebugFile(flow.Config.GetCacheDir())
	if err != nil {
		return err
	}
	defer debugFile.Close()

	files, err := flow.OpenDirFS(flow.Config.GetWorkDir())
	if err != nil {
		return err
	}
	defer files.Close()

	claude, err := flow.NewClaude(os.Getenv("ANTHROPIC_API_KEY"), flow.DefaultModel)
	if err != nil {
		return err
	}

	debug := flow.NewDebug(io.MultiWriter(os.Stdout, debugFile))
	sp := flow.NewFlowSP(
		debug,
		flow.NewTaskRepository(tasks...),
		flow.NewTaskSP(debug, claude, flow.NewFileTools(files, debug)),
	)

	return flow.NewFlow[flow.Data](sp, flow.NewStore(flow.Data{}), flow.LoadTasks{}, flow.RunTasks{Task: flow.AITask}).Run(ctx)
}
