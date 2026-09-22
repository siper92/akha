package main

import (
	"errors"
	"fmt"
	"os"
)

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("exit %d", e.code)
	}
	return e.err.Error()
}

func (e *exitError) Unwrap() error { return e.err }

func main() {
	err := newRoot().Execute()
	if err == nil {
		return
	}
	var ee *exitError
	if errors.As(err, &ee) {
		if ee.err != nil {
			fmt.Fprintln(os.Stderr, "error:", ee.err)
		}
		os.Exit(ee.code)
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
