package flow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ---------------------------------------------------------------------------
// Definition: an immutable, reusable description of a flow (ordered steps).
// Instance:   one execution of a Definition — this is the state machine.
// ---------------------------------------------------------------------------

type Definition struct {
	name     string
	steps    []Step
	index    map[string]int
	maxSteps int
}

// Builder assembles a Definition.
type Builder struct {
	def  Definition
	errs []error
}

func New(name string) *Builder {
	return &Builder{def: Definition{name: name, index: map[string]int{}, maxSteps: 1000}}
}

// Then appends steps in execution order.
func (b *Builder) Then(steps ...Step) *Builder {
	for _, s := range steps {
		if _, dup := b.def.index[s.Name()]; dup {
			b.errs = append(b.errs, fmt.Errorf("duplicate step name %q", s.Name()))
			continue
		}
		b.def.index[s.Name()] = len(b.def.steps)
		b.def.steps = append(b.def.steps, s)
	}
	return b
}

// MaxSteps guards against infinite loops created by Result.Next.
func (b *Builder) MaxSteps(n int) *Builder { b.def.maxSteps = n; return b }

func (b *Builder) Build() (*Definition, error) {
	if len(b.def.steps) == 0 {
		b.errs = append(b.errs, errors.New("flow has no steps"))
	}
	if err := errors.Join(b.errs...); err != nil {
		return nil, fmt.Errorf("flow %q: %w", b.def.name, err)
	}
	d := b.def
	return &d, nil
}

func (d *Definition) Name() string { return d.name }

// ---------------------------------------------------------------------------
// Events & observation
// ---------------------------------------------------------------------------

type Event struct {
	Time    time.Time
	FlowID  string
	Step    string // empty for flow-level events
	Attempt int
	Flow    FlowState
	StepSt  StepState
	Err     error
}

// Observer receives every state change (logging, metrics, persistence).
type Observer func(Event)

// StepRecord is the history entry for one executed (or skipped) step.
type StepRecord struct {
	Step     string
	State    StepState
	Attempts int
	Output   any
	Err      error
	Duration time.Duration
}

// ---------------------------------------------------------------------------
// Instance (the state machine)
// ---------------------------------------------------------------------------

var idSeq atomic.Uint64

type Instance struct {
	ID       string
	def      *Definition
	provider Provider
	data     *Data
	observe  []Observer

	mu      sync.RWMutex
	state   FlowState
	history []StepRecord
	err     error
	cancel  context.CancelFunc
}

type RunOption func(*Instance)

func WithData(d *Data) RunOption { return func(i *Instance) { i.data = d } }
func WithObserver(o Observer) RunOption {
	return func(i *Instance) { i.observe = append(i.observe, o) }
}
func WithID(id string) RunOption { return func(i *Instance) { i.ID = id } }

// NewInstance creates a flow execution bound to a service provider.
func (d *Definition) NewInstance(sp Provider, opts ...RunOption) *Instance {
	inst := &Instance{
		ID:       fmt.Sprintf("%s-%d", d.name, idSeq.Add(1)),
		def:      d,
		provider: sp,
		data:     NewData(),
		state:    FlowCreated,
	}
	for _, o := range opts {
		o(inst)
	}
	return inst
}

// Run is a shortcut for NewInstance(...).Run(ctx).
func (d *Definition) Run(ctx context.Context, sp Provider, opts ...RunOption) (*Instance, error) {
	inst := d.NewInstance(sp, opts...)
	return inst, inst.Run(ctx)
}

func (i *Instance) State() FlowState { i.mu.RLock(); defer i.mu.RUnlock(); return i.state }
func (i *Instance) Data() *Data      { return i.data }
func (i *Instance) Err() error       { i.mu.RLock(); defer i.mu.RUnlock(); return i.err }

func (i *Instance) History() []StepRecord {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return append([]StepRecord(nil), i.history...)
}

// Output returns the last successful output of a named step.
func (i *Instance) Output(step string) (any, bool) {
	h := i.History()
	for k := len(h) - 1; k >= 0; k-- {
		if h[k].Step == step && h[k].State == StepSucceeded {
			return h[k].Output, true
		}
	}
	return nil, false
}

// Cancel stops a running flow (or prevents a created one from starting).
func (i *Instance) Cancel() {
	i.mu.Lock()
	cancel, st := i.cancel, i.state
	i.mu.Unlock()
	if cancel != nil {
		cancel()
		return
	}
	if st == FlowCreated {
		_ = i.transition(FlowCancelled, context.Canceled)
	}
}

func (i *Instance) transition(to FlowState, err error) error {
	i.mu.Lock()
	from := i.state
	if !canTransition(from, to) {
		i.mu.Unlock()
		return fmt.Errorf("flow %s: illegal transition %s -> %s", i.ID, from, to)
	}
	i.state = to
	if err != nil {
		i.err = err
	}
	i.mu.Unlock()
	i.emit(Event{Flow: to, Err: err})
	return nil
}

func (i *Instance) emit(e Event) {
	e.Time, e.FlowID = time.Now(), i.ID
	if e.Step == "" {
		e.Flow = i.State()
	}
	for _, o := range i.observe {
		o(e)
	}
}

// Run drives the state machine until a terminal state is reached.
func (i *Instance) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	i.mu.Lock()
	i.cancel = cancel
	i.mu.Unlock()

	if err := i.transition(FlowRunning, nil); err != nil {
		return err
	}

	cursor, executed := 0, 0
	for cursor < len(i.def.steps) {
		if err := ctx.Err(); err != nil {
			_ = i.transition(FlowCancelled, err)
			return err
		}
		if executed >= i.def.maxSteps {
			err := fmt.Errorf("max steps (%d) exceeded", i.def.maxSteps)
			_ = i.transition(FlowFailed, err)
			return err
		}
		executed++

		step := i.def.steps[cursor]
		rec, res := i.runStep(ctx, step)

		i.mu.Lock()
		i.history = append(i.history, rec)
		i.mu.Unlock()

		if rec.State == StepFailed {
			if errors.Is(rec.Err, context.Canceled) {
				_ = i.transition(FlowCancelled, rec.Err)
				return rec.Err
			}
			if cfgOf(step).OnError == FailFlow {
				err := fmt.Errorf("step %q: %w", step.Name(), rec.Err)
				_ = i.transition(FlowFailed, err)
				return err
			}
		}

		// Decide the next state of the machine.
		switch {
		case res.Stop:
			cursor = len(i.def.steps)
		case res.Next != "":
			next, ok := i.def.index[res.Next]
			if !ok {
				err := fmt.Errorf("step %q routed to unknown step %q", step.Name(), res.Next)
				_ = i.transition(FlowFailed, err)
				return err
			}
			cursor = next
		default:
			cursor++
		}
	}
	return i.transition(FlowCompleted, nil)
}

func cfgOf(s Step) StepConfig {
	if c, ok := s.(configured); ok {
		return c.config()
	}
	return StepConfig{}
}

// runStep applies the step's policy (condition, retries, timeout).
func (i *Instance) runStep(ctx context.Context, step Step) (StepRecord, Result) {
	cfg := cfgOf(step)
	rec := StepRecord{Step: step.Name()}
	start := time.Now()
	sc := &StepContext{FlowID: i.ID, Step: step.Name(), Provider: i.provider, Data: i.data}

	if cfg.When != nil && !cfg.When(sc) {
		rec.State = StepSkipped
		i.emit(Event{Step: rec.Step, StepSt: StepSkipped})
		return rec, Result{}
	}

	var res Result
	var err error
	for attempt := 1; attempt <= cfg.Retries+1; attempt++ {
		sc.Attempt, rec.Attempts = attempt, attempt
		i.emit(Event{Step: rec.Step, Attempt: attempt, StepSt: StepRunning})

		res, err = runAttempt(ctx, step, sc, cfg.Timeout)
		if err == nil || ctx.Err() != nil {
			break
		}
		i.emit(Event{Step: rec.Step, Attempt: attempt, StepSt: StepFailed, Err: err})
		if attempt <= cfg.Retries && cfg.Backoff > 0 {
			select {
			case <-time.After(cfg.Backoff):
			case <-ctx.Done():
			}
		}
	}

	rec.Duration = time.Since(start)
	if err != nil {
		rec.State, rec.Err = StepFailed, err
		return rec, Result{}
	}
	rec.State, rec.Output = StepSucceeded, res.Output
	i.emit(Event{Step: rec.Step, Attempt: rec.Attempts, StepSt: StepSucceeded})
	return rec, res
}

func runAttempt(ctx context.Context, step Step, sc *StepContext, timeout time.Duration) (res Result, err error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	defer func() { // a panicking step must not crash the machine
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return step.Execute(ctx, sc)
}
