package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/siper92/akha/internal/tu"
)

func TestResolve(t *testing.T) {
	root := t.TempDir()
	sb, err := NewSandbox(root)
	if err != nil {
		t.Fatal(err)
	}

	// --- paths inside and outside the root
	cases := []tu.Case[string, string]{
		{
			Name:     "relative_file",
			Input:    "a.txt",
			Expected: filepath.Join(root, "a.txt"),
		},
		{
			Name:     "nested_dir",
			Input:    "a/b/c.txt",
			Expected: filepath.Join(root, "a", "b", "c.txt"),
		},
		{
			Name:     "dot_is_root",
			Input:    ".",
			Expected: root,
		},
		{
			Name:     "absolute_is_rebased",
			Input:    "/etc/passwd",
			Expected: filepath.Join(root, "etc", "passwd"),
		},
		{
			Name:     "dotdot_inside_stays",
			Input:    "a/../b.txt",
			Expected: filepath.Join(root, "b.txt"),
		},
		{
			Name:  "parent_escapes",
			Input: "../x.txt",
			Err:   ErrEscape,
		},
		{
			Name:  "deep_parent_escapes",
			Input: "a/../../x.txt",
			Err:   ErrEscape,
		},
		{
			Name:  "root_parent_escapes",
			Input: "..",
			Err:   ErrEscape,
		},
	}
	tu.Run(tu.New(t), cases, sb.Resolve, nil)
}

func TestFileSystem(t *testing.T) {
	root := t.TempDir()
	sb, err := NewSandbox(root)
	if err != nil {
		t.Fatal(err)
	}
	f := New(sb)
	ctx := context.Background()

	if err := f.WriteFile(ctx, "d/a.txt", "one"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := f.UpdateFile(ctx, "d/a.txt", "two"); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := f.ReadFile(ctx, "d/a.txt")
	if err != nil || got != "two" {
		t.Fatalf("read: want two, got %q err %v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, "d", "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	names, err := f.ListFiles(ctx, "d")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(names) != 2 || names[0] != "a.txt" || names[1] != "sub/" {
		t.Fatalf("list: unexpected %v", names)
	}

	// --- missing paths
	cases := []tu.Case[string, error]{
		{
			Name:  "update_missing",
			Input: "nope.txt",
			Err:   ErrNotExist,
		},
		{
			Name:  "read_missing",
			Input: "nope.txt",
			Err:   ErrNotExist,
		},
		{
			Name:  "list_missing",
			Input: "nope",
			Err:   ErrNotExist,
		},
	}
	ops := []func(string) error{
		func(p string) error { return f.UpdateFile(ctx, p, "x") },
		func(p string) error { _, err := f.ReadFile(ctx, p); return err },
		func(p string) error { _, err := f.ListFiles(ctx, p); return err },
	}
	for i, c := range cases {
		t.Run(c.Name, func(st *testing.T) {
			if err := ops[i](c.Input); err == nil || !isErr(err, c.Err) {
				st.Fatalf("want %v, got %v", c.Err, err)
			}
		})
	}
}

func isErr(err, target error) bool {
	for e := err; e != nil; {
		if e == target {
			return true
		}
		u, ok := e.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}
