package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/siper92/akha/projects/flow_gen/internal/cfg"
)

type Logger interface {
	Log(format string, args ...any)
}

var (
	once    sync.Once
	service Logger
)

func Service() Logger {
	once.Do(func() {
		service = newFileLogger(filepath.Join(cfg.GetCacheDir(), "debug"))
	})
	return service
}

type fileLogger struct {
	mu   sync.Mutex
	path string
}

var _ Logger = (*fileLogger)(nil)

func newFileLogger(dir string) *fileLogger {
	name := fmt.Sprintf("debug-%s.txt", time.Now().Format("20060102-150405"))
	return &fileLogger{path: filepath.Join(dir, name)}
}

func (l *fileLogger) Log(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	fmt.Println(line)
}

func (l *fileLogger) store(line string) error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, line)
	return err
}
