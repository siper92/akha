package runner

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/siper92/akha/internal/tu"
)

const helloScript = `
Ak.Allow(FS...)
Ak.Setup(log="hello.log", debug="hello.debug.log")
Ak.Log("hello")
Ak.Debug("dbg", 1, "two", true)
FS.WriteFile("hello/hello.txt", "hello")
FS.UpdateFile("hello/hello.txt", "hello, world")
FS.ReadFile("hello/hello.txt")
FS.ListFiles("hello/")
Ak.Exit("done", code=0)
`

func quiet() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRun(t *testing.T) {
	root := t.TempDir()
	cacheDir := t.TempDir()
	r := New(quiet())

	// --- exit codes and error kinds
	cases := []tu.Case[string, int]{
		{
			Name:     "hello_script_exits_zero",
			Input:    helloScript,
			Expected: 0,
		},
		{
			Name:     "exit_code_is_returned",
			Input:    "Ak.Allow()\nAk.Exit(\"bye\", code=3)\n",
			Expected: 3,
		},
		{
			Name:     "no_exit_call_exits_zero",
			Input:    "Ak.Allow()\nAk.Log(\"x\")\n",
			Expected: 0,
		},
		{
			Name:  "missing_allow_fails_check",
			Input: "Ak.Log(\"x\")\n",
			Err:   ErrCheck,
		},
		{
			Name:  "module_not_allowed_fails_check",
			Input: "Ak.Allow()\nFS.ListFiles(\".\")\n",
			Err:   ErrCheck,
		},
		{
			Name:  "unknown_function_fails_check",
			Input: "Ak.Allow()\nAk.Nope()\n",
			Err:   ErrCheck,
		},
		{
			Name:  "parse_error",
			Input: "Ak.Allow(\nAk.Log(\"x\")\n",
			Err:   ErrParse,
		},
		{
			Name:  "missing_file_is_runtime_error",
			Input: "Ak.Allow(FS...)\nFS.ReadFile(\"nope.txt\")\n",
			Err:   ErrRuntime,
		},
		{
			Name:  "escape_root_is_runtime_error",
			Input: "Ak.Allow(FS...)\nFS.WriteFile(\"../escape.txt\", \"x\")\n",
			Err:   ErrRuntime,
		},
		{
			Name:  "update_missing_file_is_runtime_error",
			Input: "Ak.Allow(FS...)\nFS.UpdateFile(\"missing.txt\", \"x\")\n",
			Err:   ErrRuntime,
		},
	}
	fn := func(src string) (int, error) {
		res, err := r.Run(context.Background(), src, Options{RunID: "test", Root: root, CacheDir: cacheDir})
		return res.ExitCode, err
	}
	tu.Run(tu.New(t), cases, fn, nil)
}

func TestRunWritesFiles(t *testing.T) {
	root := t.TempDir()
	cacheDir := t.TempDir()
	res, err := New(quiet()).Run(context.Background(), helloScript, Options{RunID: "r1", Root: root, CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code: want 0, got %d", res.ExitCode)
	}
	b, err := os.ReadFile(filepath.Join(root, "hello", "hello.txt"))
	if err != nil {
		t.Fatalf("read hello.txt: %v", err)
	}
	if string(b) != "hello, world" {
		t.Fatalf("hello.txt: want %q, got %q", "hello, world", string(b))
	}
	wantLog := filepath.Join(cacheDir, "runs", "r1", "hello.log")
	if res.LogPath != wantLog {
		t.Fatalf("log path: want %s, got %s", wantLog, res.LogPath)
	}
	wantDebug := filepath.Join(cacheDir, "runs", "r1", "hello.debug.log")
	if res.DebugPath != wantDebug {
		t.Fatalf("debug path: want %s, got %s", wantDebug, res.DebugPath)
	}
	for _, p := range []string{res.LogPath, res.DebugPath} {
		if st, err := os.Stat(p); err != nil || st.Size() == 0 {
			t.Fatalf("expected non empty %s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "..", "escape.txt")); err == nil {
		t.Fatal("escape.txt must not exist outside root")
	}
}

func TestRunDefaultLogPaths(t *testing.T) {
	cacheDir := t.TempDir()
	res, err := New(quiet()).Run(context.Background(), "Ak.Allow()\nAk.Log(\"x\")\n", Options{RunID: "r2", Root: t.TempDir(), CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.LogPath != filepath.Join(cacheDir, "runs", "r2", "info.log") {
		t.Fatalf("unexpected log path %s", res.LogPath)
	}
	if res.DebugPath != filepath.Join(cacheDir, "runs", "r2", "debug.log") {
		t.Fatalf("unexpected debug path %s", res.DebugPath)
	}
}

func TestCheck(t *testing.T) {
	r := New(quiet())

	// --- diagnostics count
	cases := []tu.Case[string, int]{
		{
			Name:     "valid_script",
			Input:    helloScript,
			Expected: 0,
		},
		{
			Name:     "first_call_not_allow",
			Input:    "Ak.Log(\"x\")\n",
			Expected: 1,
		},
		{
			Name:     "unknown_module",
			Input:    "Ak.Allow(HTTP...)\n",
			Expected: 1,
		},
		{
			Name:     "bad_kwarg",
			Input:    "Ak.Allow()\nAk.Exit(\"x\", status=1)\n",
			Expected: 1,
		},
		{
			Name:     "parse_error_is_reported",
			Input:    "Ak.Allow(\n",
			Expected: 1,
		},
	}
	fn := func(src string) (int, error) {
		diags, err := r.Check(context.Background(), src)
		return len(diags), err
	}
	tu.Run(tu.New(t), cases, fn, nil)
}
