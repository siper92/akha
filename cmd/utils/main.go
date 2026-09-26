package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/siper92/akha/cmd/utils/sourcemap"
)

var ErrUsage = errors.New("usage: akha_utils source-map --path <path> --output <output_path> [--with-private]")

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}
	switch args[0] {
	case "source-map":
		return runSourceMap(ctx, args[1:])
	}
	return fmt.Errorf("%w: unknown command %q", ErrUsage, args[0])
}

func runSourceMap(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("source-map", flag.ContinueOnError)
	path := fs.String("path", "", "directory to parse recursively")
	output := fs.String("output", "", "directory for the source map file")
	withPrivate := fs.Bool("with-private", false, "include private funcs and types")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" || *output == "" {
		return ErrUsage
	}
	abs, err := filepath.Abs(*path)
	if err != nil {
		return err
	}
	files, err := sourcemap.NewParser().Parse(ctx, abs, sourcemap.Options{WithPrivate: *withPrivate})
	if err != nil {
		return err
	}
	data, err := sourcemap.NewEncoder().Encode(ctx, files)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*output, 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("source_map_%s_%s.goyml", filepath.Base(abs), time.Now().Format("20060102_150405"))
	target := filepath.Join(*output, name)
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return err
	}
	fmt.Println(target)
	return nil
}
