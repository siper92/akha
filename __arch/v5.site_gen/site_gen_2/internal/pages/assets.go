package pages

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type dirAssets struct {
	src string
	dst string
}

var _ Assets = (*dirAssets)(nil)

func NewDirAssets(src, dst string) Assets {
	return &dirAssets{src: src, dst: dst}
}

func (a *dirAssets) Copy(ctx context.Context) error {
	return filepath.WalkDir(a.src, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if strings.HasPrefix(e.Name(), ".") && path != a.src {
			if e.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(a.src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(a.dst, rel)
		if e.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if filepath.Ext(e.Name()) == ".md" {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copy %s: %w", src, err)
	}
	return out.Close()
}

func Build(ctx context.Context, docs []Doc, r Renderer, w Writer, a Assets) ([]Page, error) {
	built := make([]Page, 0, len(docs))
	for _, d := range docs {
		p, err := r.Render(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", d.Slug, err)
		}
		built = append(built, p)
	}
	if err := w.Write(ctx, built); err != nil {
		return nil, err
	}
	if err := a.Copy(ctx); err != nil {
		return nil, fmt.Errorf("copy assets: %w", err)
	}
	return built, nil
}
