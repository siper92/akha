package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/siper92/akha/lang/runner"
	"github.com/siper92/akha/platform/worker/client"
)

var (
	ErrNotStarted = errors.New("worker not started")
	ErrAuth       = errors.New("worker auth failed")
)

type service struct {
	ts    client.TokenSource
	be    client.Backend
	run   runner.Runner
	log   *slog.Logger
	token string
}

var _ Worker = (*service)(nil)

func New(ts client.TokenSource, be client.Backend, run runner.Runner, log *slog.Logger) Worker {
	if log == nil {
		log = slog.Default()
	}
	return &service{ts: ts, be: be, run: run, log: log}
}

func (s *service) Start(ctx context.Context) error {
	token, err := s.authenticate(ctx)
	if err != nil {
		return err
	}
	s.token = token
	s.log.Info("worker ready")
	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.token = ""
	s.log.Info("worker stopped")
	return nil
}

func (s *service) Run(ctx context.Context, src string, opts runner.Options) (runner.Result, error) {
	if s.token == "" {
		return runner.Result{ExitCode: runner.ExitRuntime}, ErrNotStarted
	}
	return s.run.Run(ctx, src, opts)
}

func (s *service) authenticate(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, err := s.ts.Token(ctx)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrAuth, err)
		}
		ok, err := s.be.Validate(ctx, token)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrAuth, err)
		}
		if ok {
			return token, nil
		}
		s.log.Warn("cached token rejected, logging in again")
		if err := s.ts.Invalidate(ctx); err != nil {
			return "", fmt.Errorf("%w: %w", ErrAuth, err)
		}
	}
	return "", fmt.Errorf("%w: token rejected", ErrAuth)
}
