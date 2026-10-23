package state_machine

import (
	"context"
	"errors"
)

var errBoom = errors.New("boom")

type counter struct {
	N, Limit int
	Path     []string
}

type result struct {
	State string
	Path  []string
}

type runInput struct {
	m    *Machine[counter, string]
	ctx  context.Context
	from string
	data counter
}

func visit(next string) StepFunc[counter, string] {
	return func(_ context.Context, c *counter) (string, error) {
		c.Path = append(c.Path, next)
		return next, nil
	}
}

func work(_ context.Context, c *counter) (string, error) {
	c.N++
	c.Path = append(c.Path, "work")
	if c.N >= c.Limit {
		return "done", nil
	}
	return "work", nil
}

func fail(_ context.Context, _ *counter) (string, error) {
	return "", errBoom
}

func explode(_ context.Context, _ *counter) (string, error) {
	panic("kaboom")
}

func fixture() *Machine[counter, string] {
	return NewMachine[counter, string]("load").
		Step("load", visit("work"), "work").
		Step("work", work, "work", "done").
		Final("done")
}

func canceled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
