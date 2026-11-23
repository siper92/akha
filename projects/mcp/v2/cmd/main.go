package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/anthropics/anthropic-sdk-go/v2"

	akmcp "github.com/siper92/akha/projects/mcp/v2"
)

const defaultPrompt = "Create a short story of 3 parts in 3 files: file1.txt, file2.txt and file3.txt. Use the createFile tool once per part."

func main() {
	dir := flag.String("dir", "./out", "directory the createFile tool writes into")
	prompt := flag.String("prompt", defaultPrompt, "prompt sent to Claude")
	model := flag.String("model", akmcp.DefaultModel, "Claude model id")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tools, err := akmcp.NewService(akmcp.NewCreateFile(*dir), akmcp.NewReadFile(*dir))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	agent := akmcp.NewClaudeAgent(anthropic.NewClient(), tools, *model, log)

	out, err := agent.Run(ctx, *prompt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Println(out)
}
