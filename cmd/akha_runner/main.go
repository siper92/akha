package main

import (
	"fmt"
	"os"

	"github.com/siper92/akha/cmd/akha_runner/cmd"
)

func main() {
	err := cmd.NewRoot().Execute()
	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
