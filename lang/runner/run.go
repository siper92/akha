package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/check"
	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/module/ak"
	"github.com/siper92/akha/lang/module/fs"
	"github.com/siper92/akha/lang/parser"
)

const (
	ExitOK      = 0
	ExitRuntime = 1
	ExitCheck   = 2

	DefaultCacheDir = "_env/.cache/akha"
	DefaultRoot     = "."
)

var (
	ErrParse   = errors.New("parse failed")
	ErrCheck   = errors.New("check failed")
	ErrRuntime = errors.New("runtime error")
)

type runner struct {
	log *slog.Logger
}

var _ Runner = (*runner)(nil)

func New(log *slog.Logger) Runner {
	if log == nil {
		log = slog.Default()
	}
	return &runner{log: log}
}

func (r *runner) Check(ctx context.Context, src string) ([]check.Diagnostic, error) {
	s, diags := parse(src)
	if len(diags) > 0 {
		return diags, nil
	}
	reg, out, err := r.registry(DefaultRoot, os.TempDir())
	if err != nil {
		return nil, err
	}
	defer out.Close()
	return check.New().Check(ctx, s, reg), nil
}

func (r *runner) Run(ctx context.Context, src string, opts Options) (Result, error) {
	opts = withDefaults(opts)
	runDir := filepath.Join(opts.CacheDir, "runs", opts.RunID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return Result{ExitCode: ExitRuntime}, err
	}
	s, diags := parse(src)
	if len(diags) > 0 {
		return Result{ExitCode: ExitCheck, Diagnostics: diags}, ErrParse
	}
	reg, out, err := r.registry(opts.Root, runDir)
	if err != nil {
		return Result{ExitCode: ExitRuntime}, err
	}
	defer out.Close()
	res := Result{LogPath: out.LogPath(), DebugPath: out.DebugPath()}
	if diags = check.New().Check(ctx, s, reg); len(diags) > 0 {
		res.ExitCode = ExitCheck
		res.Diagnostics = diags
		return res, ErrCheck
	}
	r.log.Info("run start", "run", opts.RunID, "root", opts.Root)
	err = eval.New(reg).Eval(ctx, s)
	res.LogPath = out.LogPath()
	res.DebugPath = out.DebugPath()
	var exit *eval.ExitError
	switch {
	case errors.As(err, &exit):
		res.ExitCode = exit.Code
	case err != nil:
		res.ExitCode = ExitRuntime
		r.log.Error("run failed", "run", opts.RunID, "err", err)
		return res, fmt.Errorf("%w: %w", ErrRuntime, err)
	}
	r.log.Info("run done", "run", opts.RunID, "exit", res.ExitCode)
	return res, nil
}

func (r *runner) registry(root, runDir string) (eval.Registry, ak.FileOutput, error) {
	sb, err := fs.NewSandbox(root)
	if err != nil {
		return nil, nil, err
	}
	out := ak.NewOutput(runDir, r.log)
	reg := eval.NewRegistry()
	if err := reg.Register(ak.New(out, ak.NewAllowed())); err != nil {
		return nil, nil, err
	}
	if err := reg.Register(fs.NewModule(fs.New(sb))); err != nil {
		return nil, nil, err
	}
	return reg, out, nil
}

func parse(src string) (*ast.Script, []check.Diagnostic) {
	s, err := parser.New(lexer.New(src)).Parse()
	var errs parser.Errors
	if errors.As(err, &errs) {
		diags := make([]check.Diagnostic, len(errs))
		for i, e := range errs {
			diags[i] = check.Diagnostic{Pos: e.Pos, Severity: check.SeverityError, Msg: e.Msg}
		}
		return s, diags
	}
	if err != nil {
		return s, []check.Diagnostic{{Severity: check.SeverityError, Msg: err.Error()}}
	}
	return s, nil
}

func withDefaults(o Options) Options {
	if o.RunID == "" {
		o.RunID = time.Now().UTC().Format("20060102-150405.000000")
	}
	if o.CacheDir == "" {
		o.CacheDir = DefaultCacheDir
	}
	if o.Root == "" {
		o.Root = DefaultRoot
	}
	return o
}
