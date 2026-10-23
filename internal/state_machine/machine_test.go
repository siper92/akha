package state_machine

import (
	"context"
	"fmt"
	"testing"

	tu "github.com/siper92/akha/lang/tests_utils"
)

// --- transitions from injected state

func TestRunFrom_Transitions(t *testing.T) {
	cases := []tu.Case[runInput, result]{
		{
			Name:     "start_state_loops_until_limit",
			Input:    runInput{from: "load", data: counter{Limit: 2}},
			Expected: result{State: "done", Path: []string{"work", "work", "work"}},
		},
		{
			Name:     "inject_work_state",
			Input:    runInput{from: "work", data: counter{Limit: 1}},
			Expected: result{State: "done", Path: []string{"work"}},
		},
		{
			Name:     "inject_partial_data",
			Input:    runInput{from: "work", data: counter{N: 1, Limit: 3}},
			Expected: result{State: "done", Path: []string{"work", "work"}},
		},
		{
			Name:     "inject_final_state_is_noop",
			Input:    runInput{from: "done", data: counter{Limit: 5}},
			Expected: result{State: "done"},
		},
	}

	tu.Run(tu.New(t), cases, func(in runInput) (result, error) {
		state, err := fixture().RunFrom(context.Background(), in.from, &in.data)
		return result{State: state, Path: in.data.Path}, err
	}, nil)
}

// --- run errors

func TestRunFrom_Errors(t *testing.T) {
	cases := []tu.Case[runInput, result]{
		{
			Name:  "unknown_injected_state",
			Input: runInput{m: fixture(), from: "nope"},
			Err:   ErrUnknownState,
		},
		{
			Name:  "step_error_is_wrapped",
			Input: runInput{m: NewMachine[counter, string]("a").Step("a", fail, "done").Final("done"), from: "a"},
			Err:   errBoom,
		},
		{
			Name:  "step_panic_is_recovered",
			Input: runInput{m: NewMachine[counter, string]("a").Step("a", explode, "done").Final("done"), from: "a"},
			Err:   ErrStepPanic,
		},
		{
			Name:  "transition_not_allowed",
			Input: runInput{m: NewMachine[counter, string]("a").Step("a", visit("done"), "b").Step("b", visit("done"), "done").Final("done"), from: "a"},
			Err:   ErrInvalidTransition,
		},
		{
			Name:  "endless_loop_hits_max_steps",
			Input: runInput{m: NewMachine[counter, string]("a").Step("a", visit("a"), "a").MaxSteps(3), from: "a"},
			Err:   ErrMaxSteps,
		},
		{
			Name:  "canceled_context",
			Input: runInput{m: fixture(), ctx: canceled(), from: "load"},
			Err:   context.Canceled,
		},
	}

	tu.Run(tu.New(t), cases, func(in runInput) (result, error) {
		ctx := in.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		state, err := in.m.RunFrom(ctx, in.from, &in.data)
		return result{State: state, Path: in.data.Path}, err
	}, nil)
}

// --- definition validation

func TestValidate(t *testing.T) {
	cases := []tu.Case[*Machine[counter, string], bool]{
		{
			Name:     "valid_definition",
			Input:    fixture(),
			Expected: true,
		},
		{
			Name:  "duplicate_step",
			Input: fixture().Step("load", visit("work"), "work"),
			Err:   ErrDuplicateStep,
		},
		{
			Name:  "nil_step_func",
			Input: fixture().Step("idle", nil),
			Err:   ErrNilStep,
		},
		{
			Name:  "unknown_target",
			Input: fixture().Step("idle", visit("missing"), "missing"),
			Err:   ErrUnknownState,
		},
		{
			Name:  "unknown_start",
			Input: NewMachine[counter, string]("missing").Final("done"),
			Err:   ErrUnknownState,
		},
	}

	tu.Run(tu.New(t), cases, func(m *Machine[counter, string]) (bool, error) {
		err := m.Validate()
		return err == nil, err
	}, nil)
}

// --- step hook

func TestOnStep_Trace(t *testing.T) {
	cases := []tu.Case[runInput, []string]{
		{
			Name:     "trace_from_start",
			Input:    runInput{m: fixture(), from: "load", data: counter{Limit: 1}},
			Expected: []string{"load>work", "work>done"},
		},
		{
			Name:     "trace_records_failure",
			Input:    runInput{m: NewMachine[counter, string]("a").Step("a", fail, "done").Final("done"), from: "a"},
			Expected: []string{"a>:err"},
		},
		{
			Name:     "trace_empty_on_final",
			Input:    runInput{m: fixture(), from: "done"},
			Expected: nil,
		},
	}

	tu.Run(tu.New(t), cases, func(in runInput) ([]string, error) {
		var trace []string
		in.m.OnStep = func(tr Transition[string]) {
			s := fmt.Sprintf("%s>%s", tr.From, tr.To)
			if tr.Err != nil {
				s += ":err"
			}
			trace = append(trace, s)
		}
		_, _ = in.m.RunFrom(context.Background(), in.from, &in.data)
		return trace, nil
	}, nil)
}
