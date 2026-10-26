package flow

import (
	"context"
	"fmt"
	"sync"

	"github.com/siper92/akha/internal/data_sources"
)

var _ Services = ServiceMap(nil)

type ServiceMap map[string]any

func (m ServiceMap) Service(name string) (any, bool) {
	s, ok := m[name]
	return s, ok
}

var _ Flow = (*Runner)(nil)

type Runner struct {
	plan     Plan
	steps    map[ID]Step
	sources  data_sources.Provider
	services Services
	proc     StepProcessor

	mu      sync.RWMutex
	results map[ID]any
	cur     int
}

func New(plan Plan, sources data_sources.Provider, services Services, steps ...Step) (*Runner, error) {
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	byID := make(map[ID]Step, len(steps))
	for _, s := range steps {
		if _, ok := byID[s.ID()]; ok {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateID, s.ID())
		}
		if _, ok := plan.Index(s.ID()); !ok {
			return nil, fmt.Errorf("%w: %s not in plan", ErrUnknownStep, s.ID())
		}
		byID[s.ID()] = s
	}
	for _, id := range plan.Order() {
		if _, ok := byID[id]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownStep, id)
		}
	}
	return &Runner{
		plan:     plan,
		steps:    byID,
		sources:  sources,
		services: services,
		proc:     ProcessorFunc[StepRun, any](processStep),
		results:  map[ID]any{},
		cur:      -1,
	}, nil
}

func (r *Runner) Use(p StepProcessor) *Runner {
	r.proc = p
	return r
}

func (r *Runner) Plan() Plan { return r.plan }

func (r *Runner) Sources() data_sources.Provider { return r.sources }

func (r *Runner) Services() Services { return r.services }

func (r *Runner) Result(id ID) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.results[id]
	return v, ok
}

func (r *Runner) Call(ctx context.Context, id ID) (any, error) {
	return r.call(ctx, r.cur, id)
}

func (r *Runner) Exec(ctx context.Context) error {
	for i, id := range r.plan.Order() {
		r.cur = i
		if _, err := r.run(ctx, i, id); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) call(ctx context.Context, cur int, id ID) (any, error) {
	i, ok := r.plan.Index(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownStep, id)
	}
	if i >= cur {
		return nil, fmt.Errorf("%w: %s", ErrNotPrevious, id)
	}
	return r.run(ctx, i, id)
}

func (r *Runner) run(ctx context.Context, i int, id ID) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out, err := r.proc.Process(ctx, StepRun{Step: r.steps[id], Scope: &stepScope{r: r, idx: i}})
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrStepFailed, id, err)
	}
	r.mu.Lock()
	r.results[id] = out
	r.mu.Unlock()
	return out, nil
}

var _ StepScope = (*stepScope)(nil)

type stepScope struct {
	r   *Runner
	idx int
}

func (s *stepScope) Sources() data_sources.Provider { return s.r.sources }

func (s *stepScope) Services() Services { return s.r.services }

func (s *stepScope) Result(id ID) (any, bool) { return s.r.Result(id) }

func (s *stepScope) Call(ctx context.Context, id ID) (any, error) {
	return s.r.call(ctx, s.idx, id)
}
