package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

var ErrSpawn = errors.New("spawn failed")

type Factory func(ctx context.Context, i int) (Worker, error)

type spawner struct {
	factory Factory
	log     *slog.Logger
	wg      sync.WaitGroup
}

var _ Spawner = (*spawner)(nil)

func NewSpawner(f Factory, log *slog.Logger) Spawner {
	if log == nil {
		log = slog.Default()
	}
	return &spawner{factory: f, log: log}
}

func (s *spawner) Spawn(ctx context.Context, n int) error {
	for i := 0; i < n; i++ {
		w, err := s.factory(ctx, i)
		if err != nil {
			return fmt.Errorf("%w: worker %d: %w", ErrSpawn, i, err)
		}
		if err := w.Start(ctx); err != nil {
			return fmt.Errorf("%w: worker %d: %w", ErrSpawn, i, err)
		}
		s.wg.Add(1)
		go func(i int, w Worker) {
			defer s.wg.Done()
			<-ctx.Done()
			if err := w.Stop(context.WithoutCancel(ctx)); err != nil {
				s.log.Error("embedded worker stop", "worker", i, "err", err)
			}
		}(i, w)
		s.log.Info("embedded worker started", "worker", i)
	}
	return nil
}
