package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	akmcp "github.com/siper92/akha/projects/mcp/v1"
)

const defaultPrompt = "Create a short story of 3 parts in 3 files: part1.txt, part2.txt and part3.txt. Use the createFile tool once per part."

func main() {
	root := flag.String("root", "./out", "directory the createFile tool writes into")
	prompt := flag.String("prompt", defaultPrompt, "prompt sent to Claude")
	model := flag.String("model", akmcp.DefaultModel, "Claude model id")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, *root, *prompt, *model); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, root, prompt, model string) error {
	server := akmcp.NewServer(root)

	serverT, clientT := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, serverT, nil)
	if err != nil {
		return fmt.Errorf("server connect: %w", err)
	}
	defer ss.Close()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "akha-agent",
		Version: akmcp.ServerVersion,
	}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		return fmt.Errorf("client connect: %w", err)
	}
	defer cs.Close()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	agent := akmcp.NewClaudeAgent(anthropic.NewClient(), cs, model, log)
	out, err := agent.Run(ctx, prompt)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}
