package ak

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/siper92/akha/lang/eval"
)

const (
	DefaultLogName   = "info.log"
	DefaultDebugName = "debug.log"
)

var ErrOutputClosed = errors.New("output closed")

type FileOutput interface {
	Output
	LogPath() string
	DebugPath() string
}

type fileOutput struct {
	dir       string
	log       *slog.Logger
	logPath   string
	debugPath string
	logF      *os.File
	debugF    *os.File
	closed    bool
}

var _ FileOutput = (*fileOutput)(nil)

func NewOutput(dir string, log *slog.Logger) FileOutput {
	if log == nil {
		log = slog.Default()
	}
	return &fileOutput{
		dir:       dir,
		log:       log,
		logPath:   filepath.Join(dir, DefaultLogName),
		debugPath: filepath.Join(dir, DefaultDebugName),
	}
}

func (o *fileOutput) LogPath() string   { return o.logPath }
func (o *fileOutput) DebugPath() string { return o.debugPath }

func (o *fileOutput) Setup(ctx context.Context, logPath, debugPath string) error {
	if o.closed {
		return ErrOutputClosed
	}
	if logPath != "" {
		if err := closeFile(&o.logF); err != nil {
			return err
		}
		o.logPath = o.resolve(logPath)
	}
	if debugPath != "" {
		if err := closeFile(&o.debugF); err != nil {
			return err
		}
		o.debugPath = o.resolve(debugPath)
	}
	return nil
}

func (o *fileOutput) Log(ctx context.Context, msg string) error {
	if o.closed {
		return ErrOutputClosed
	}
	o.log.Info(msg, "src", "ak")
	f, err := o.file(&o.logF, o.logPath)
	if err != nil {
		return err
	}
	return writeLine(f, msg)
}

func (o *fileOutput) Debug(ctx context.Context, msg string, args ...eval.Value) error {
	if o.closed {
		return ErrOutputClosed
	}
	line := msg + formatArgs(args)
	o.log.Debug(line, "src", "ak")
	f, err := o.file(&o.debugF, o.debugPath)
	if err != nil {
		return err
	}
	return writeLine(f, line)
}

func (o *fileOutput) Close() error {
	o.closed = true
	return errors.Join(closeFile(&o.logF), closeFile(&o.debugF))
}

func (o *fileOutput) resolve(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(o.dir, p)
}

func (o *fileOutput) file(slot **os.File, path string) (*os.File, error) {
	if *slot != nil {
		return *slot, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	*slot = f
	return f, nil
}

func closeFile(slot **os.File) error {
	if *slot == nil {
		return nil
	}
	err := (*slot).Close()
	*slot = nil
	return err
}

func writeLine(f *os.File, line string) error {
	_, err := fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339), line)
	return err
}
