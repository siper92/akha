package pages

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

type dirWriter struct {
	out     string
	baseURL string
}

var _ Writer = (*dirWriter)(nil)

func NewDirWriter(out, baseURL string) Writer {
	return &dirWriter{out: out, baseURL: baseURL}
}

func (w *dirWriter) Reset(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	abs, err := filepath.Abs(w.out)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnsafeOut, err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnsafeOut, err)
	}
	home, _ := os.UserHomeDir()
	if abs == filepath.Dir(abs) || abs == cwd || abs == home {
		return fmt.Errorf("%w: %s", ErrUnsafeOut, abs)
	}
	if err := os.RemoveAll(abs); err != nil {
		return fmt.Errorf("clean %s: %w", abs, err)
	}
	return nil
}

func (w *dirWriter) Write(ctx context.Context, pages []Page) error {
	for _, p := range pages {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.writeFile(p.File, p.HTML); err != nil {
			return err
		}
	}
	if w.baseURL == "" {
		return nil
	}
	return w.writeFile("sitemap.xml", w.sitemap(pages))
}

func (w *dirWriter) writeFile(rel string, data []byte) error {
	path := filepath.Join(w.out, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func (w *dirWriter) sitemap(pages []Page) []byte {
	var b bytes.Buffer
	b.WriteString(xml.Header)
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, p := range pages {
		b.WriteString("  <url><loc>")
		_ = xml.EscapeText(&b, []byte(w.baseURL+p.Path))
		b.WriteString("</loc></url>\n")
	}
	b.WriteString("</urlset>\n")
	return b.Bytes()
}
