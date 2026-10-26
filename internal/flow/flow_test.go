package flow

import (
	"context"
	"testing"

	"github.com/siper92/akha/internal/data_sources"
	tu "github.com/siper92/akha/lang/tests_utils"
)

// --- flow executes steps, steps run task chains per data item

func TestExec(t *testing.T) {
	cases := []tu.Case[flowInput, int]{
		{
			Name:     "sequential_items",
			Input:    flowInput{nums: []int{1, 2, 3}, source: "nums"},
			Expected: 15,
		},
		{
			Name:     "parallel_items",
			Input:    flowInput{nums: []int{1, 2, 3, 4}, workers: 3, source: "nums"},
			Expected: 24,
		},
		{
			Name:     "empty_source",
			Input:    flowInput{source: "nums"},
			Expected: 0,
		},
		{
			Name:  "unknown_source",
			Input: flowInput{source: "missing"},
			Err:   data_sources.ErrSourceNotFound,
		},
	}

	tu.Run(tu.New(t), cases, func(in flowInput) (int, error) {
		f, err := newFlow(in)
		if err != nil {
			return 0, err
		}
		if err := f.Exec(context.Background()); err != nil {
			return 0, err
		}
		out, err := ResultAs[[]int](f, "sum")
		if err != nil {
			return 0, err
		}
		return out[0], nil
	}, nil)
}

// --- calls are limited to previous steps

func TestCall_NotPrevious(t *testing.T) {
	cases := []tu.Case[ID, any]{
		{
			Name:  "call_before_exec",
			Input: "double",
			Err:   ErrNotPrevious,
		},
		{
			Name:  "unknown_step",
			Input: "nope",
			Err:   ErrUnknownStep,
		},
	}

	tu.Run(tu.New(t), cases, func(id ID) (any, error) {
		f, err := newFlow(flowInput{source: "nums"})
		if err != nil {
			return nil, err
		}
		return f.Call(context.Background(), id)
	}, nil)
}
