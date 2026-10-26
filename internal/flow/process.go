package flow

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type ProcessorFunc[E, R any] func(ctx context.Context, e E) (R, error)

func (f ProcessorFunc[E, R]) Process(ctx context.Context, e E) (R, error) { return f(ctx, e) }

func processStep(ctx context.Context, r StepRun) (any, error) {
	return r.Step.Process(ctx, r.Scope)
}

func processTask[D, R any](ctx context.Context, r TaskRun[D, R]) (R, error) {
	return r.Task.Run(ctx, r.Data, r.Scope)
}

func ProcessAll[E, R any](ctx context.Context, p Processor[E, R], items []E, workers int) ([]R, error) {
	out := make([]R, len(items))
	errs := make([]error, len(items))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	idx := make(chan int)
	var wg sync.WaitGroup
	for range max(1, min(workers, len(items))) {
		wg.Go(func() {
			for i := range idx {
				r, err := p.Process(ctx, items[i])
				if err != nil {
					errs[i] = fmt.Errorf("item %d: %w", i, err)
					cancel()
					continue
				}
				out[i] = r
			}
		})
	}

dispatch:
	for i := range items {
		select {
		case idx <- i:
		case <-ctx.Done():
			errs[i] = ctx.Err()
			break dispatch
		}
	}
	close(idx)
	wg.Wait()

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return out, nil
}
