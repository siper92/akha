package v3

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var _ Debug = (*WriterDebug)(nil)

type WriterDebug struct {
	out io.Writer
}

func NewDebug(out io.Writer) *WriterDebug {
	return &WriterDebug{out: out}
}

func (d *WriterDebug) Log(format string, args ...any) {
	fmt.Fprintln(d.out, fmt.Sprintf(format, args...))
}

func OpenDebugFile(cacheDir string) (*os.File, error) {
	dir := filepath.Join(cacheDir, "debug")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	name := "debug-" + time.Now().Format("20060102-150405") + ".txt"

	return os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
