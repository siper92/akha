// A simple state machine with branching.
//
// Example flow:
//
//	            ┌── invalid ──────────────────────────────► reject (final)
//	validate ───┤
//	            └── valid ──► check_stock ──┬── in_stock ──► charge ──┬── paid ──────► ship (final)
//	                                        │                         ├── declined ──► charge  (retry, while attempts < 3)
//	                                        │                         └── declined ──► cancel (final, otherwise)
//	                                        └── out_of_stock ─► backorder (final)
//
// A step can have several possible next steps. Each one is a Transition with:
//   - On:    the outcome the step reported ("" means any outcome)
//   - When:  an optional guard that can inspect flow data
//   - To:    the next step
//
// The engine checks the outgoing transitions in the order they were added
// and takes the first one that matches.
package main

import (
	"context"
	"errors"
	"fmt"
)

// ---------------------------------------------------------------------
// Contracts
// ---------------------------------------------------------------------

type StepID string
type Outcome string

// Data is the shared state of one run.
type Data map[string]any

// Services is what the machine hands to each step.
type Services interface {
	Data() Data
}

// Step does its work and reports an outcome. It does NOT choose the next
// step; the transitions do.
type Step interface {
	ID() StepID
	Execute(ctx context.Context, sp Services) (Outcome, error)
}

// Guard is an extra condition on a transition.
type Guard func(sp Services) bool

// Transition is one possible edge out of a step.
type Transition struct {
	From StepID
	On   Outcome // "" = any outcome
	When Guard   // nil = always
	To   StepID
}

func (t Transition) matches(o Outcome, sp Services) bool {
	if t.On != "" && t.On != o {
		return false
	}
	return t.When == nil || t.When(sp)
}

// ---------------------------------------------------------------------
// Machine
// ---------------------------------------------------------------------

type Machine struct {
	start  StepID
	final  map[StepID]bool
	steps  map[StepID]Step
	routes map[StepID][]Transition // several next steps per step
}

func New(start StepID) *Machine {
	return &Machine{
		start:  start,
		final:  map[StepID]bool{},
		steps:  map[StepID]Step{},
		routes: map[StepID][]Transition{},
	}
}

func (m *Machine) Step(s Step) *Machine { m.steps[s.ID()] = s; return m }

func (m *Machine) Final(ids ...StepID) *Machine {
	for _, id := range ids {
		m.final[id] = true
	}
	return m
}

// On adds a branch chosen only by the outcome.
func (m *Machine) On(from StepID, o Outcome, to StepID) *Machine {
	return m.Add(Transition{From: from, On: o, To: to})
}

// OnIf adds a branch chosen by the outcome and a guard.
func (m *Machine) OnIf(from StepID, o Outcome, g Guard, to StepID) *Machine {
	return m.Add(Transition{From: from, On: o, When: g, To: to})
}

func (m *Machine) Add(t Transition) *Machine {
	m.routes[t.From] = append(m.routes[t.From], t)
	return m
}

// Validate checks the graph before running it.
func (m *Machine) Validate() error {
	if _, ok := m.steps[m.start]; !ok {
		return fmt.Errorf("start step %q not registered", m.start)
	}
	for from, ts := range m.routes {
		if _, ok := m.steps[from]; !ok {
			return fmt.Errorf("transition from unknown step %q", from)
		}
		for _, t := range ts {
			if _, ok := m.steps[t.To]; !ok {
				return fmt.Errorf("transition %s -> unknown step %q", from, t.To)
			}
		}
	}
	for id := range m.steps {
		if !m.final[id] && len(m.routes[id]) == 0 {
			return fmt.Errorf("step %q is not final and has no transitions", id)
		}
	}
	return nil
}

var ErrNoRoute = errors.New("no transition matched")

type services struct{ data Data }

func (s services) Data() Data { return s.data }

// Run executes steps until it reaches a final step. It returns the visited path.
func (m *Machine) Run(ctx context.Context, data Data) ([]StepID, error) {
	sp := services{data: data}
	cur := m.start
	var path []StepID

	for {
		if err := ctx.Err(); err != nil {
			return path, err
		}

		path = append(path, cur)

		outcome, err := m.steps[cur].Execute(ctx, sp)
		if err != nil {
			return path, fmt.Errorf("step %s: %w", cur, err)
		}

		if m.final[cur] {
			return path, nil
		}

		next, ok := m.next(cur, outcome, sp)
		if !ok {
			return path, fmt.Errorf("step %s, outcome %q: %w", cur, outcome, ErrNoRoute)
		}

		cur = next
	}
}

// next takes the first matching transition, in the order they were added.
func (m *Machine) next(from StepID, o Outcome, sp Services) (StepID, bool) {
	for _, t := range m.routes[from] {
		if t.matches(o, sp) {
			return t.To, true
		}
	}

	return "", false
}

// ---------------------------------------------------------------------
// Steps for the example (a function adapter keeps them short)
// ---------------------------------------------------------------------

type StepFunc struct {
	id StepID
	fn func(ctx context.Context, sp Services) (Outcome, error)
}

func (s StepFunc) ID() StepID { return s.id }
func (s StepFunc) Execute(ctx context.Context, sp Services) (Outcome, error) {
	return s.fn(ctx, sp)
}

func step(id StepID, fn func(ctx context.Context, sp Services) (Outcome, error)) Step {
	return StepFunc{id: id, fn: fn}
}

func done(msg string) func(context.Context, Services) (Outcome, error) {
	return func(_ context.Context, sp Services) (Outcome, error) {
		sp.Data()["result"] = msg
		return "done", nil
	}
}

func buildOrderFlow() *Machine {
	const maxAttempts = 3

	m := New("validate").
		Step(step("validate", func(_ context.Context, sp Services) (Outcome, error) {
			if sp.Data()["qty"].(int) <= 0 {
				return "invalid", nil
			}
			return "valid", nil
		})).
		Step(step("check_stock", func(_ context.Context, sp Services) (Outcome, error) {
			if sp.Data()["qty"].(int) > sp.Data()["stock"].(int) {
				return "out_of_stock", nil
			}
			return "in_stock", nil
		})).
		Step(step("charge", func(_ context.Context, sp Services) (Outcome, error) {
			d := sp.Data()
			d["attempts"] = d["attempts"].(int) + 1
			if d["attempts"].(int) < d["succeedOnAttempt"].(int) {
				return "declined", nil
			}
			return "paid", nil
		})).
		Step(step("ship", done("shipped"))).
		Step(step("reject", done("rejected"))).
		Step(step("backorder", done("backordered"))).
		Step(step("cancel", done("cancelled: payment declined"))).
		Final("ship", "reject", "backorder", "cancel")

	// Branching: each step lists ALL of its possible next steps.
	m.On("validate", "invalid", "reject").
		On("validate", "valid", "check_stock")

	m.On("check_stock", "in_stock", "charge").
		On("check_stock", "out_of_stock", "backorder")

	// Same outcome, two destinations: the guard decides. Order matters.
	canRetry := func(sp Services) bool { return sp.Data()["attempts"].(int) < maxAttempts }
	m.On("charge", "paid", "ship").
		OnIf("charge", "declined", canRetry, "charge"). // loop back
		On("charge", "declined", "cancel")              // fallback

	return m
}

// ---------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------

func main() {
	m := buildOrderFlow()
	if err := m.Validate(); err != nil {
		panic(err)
	}

	cases := []struct {
		name string
		data Data
	}{
		{"happy path", Data{"qty": 2, "stock": 5, "attempts": 0, "succeedOnAttempt": 1}},
		{"invalid qty", Data{"qty": 0, "stock": 5, "attempts": 0, "succeedOnAttempt": 1}},
		{"out of stock", Data{"qty": 9, "stock": 5, "attempts": 0, "succeedOnAttempt": 1}},
		{"retry then pay", Data{"qty": 1, "stock": 5, "attempts": 0, "succeedOnAttempt": 2}},
		{"retries exhausted", Data{"qty": 1, "stock": 5, "attempts": 0, "succeedOnAttempt": 9}},
	}

	for _, c := range cases {
		path, err := m.Run(context.Background(), c.data)
		if err != nil {
			fmt.Printf("%-18s error: %v\n", c.name, err)
			continue
		}
		fmt.Printf("%-18s %v -> %s\n", c.name, path, c.data["result"])
	}
}
