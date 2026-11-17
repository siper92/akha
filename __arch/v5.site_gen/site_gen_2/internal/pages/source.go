package pages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const outlineSuffix = "_outline"

type dirSource struct {
	dir       string
	indexPage string
}

var _ Source = (*dirSource)(nil)

func NewDirSource(dir, indexPage string) Source {
	return &dirSource{dir: dir, indexPage: indexPage}
}

func (s *dirSource) List(ctx context.Context) ([]Doc, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", s.dir, err)
	}

	var docs []Doc
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		if strings.HasSuffix(slug, outlineSuffix) || strings.HasPrefix(slug, "_") {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		meta, body, err := splitFrontMatter(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if meta.Draft {
			continue
		}

		d := Doc{Slug: slug, Path: "/" + slug + "/", File: slug + "/index.html", Meta: meta, Body: body}
		if slug == s.indexPage {
			d.Path, d.File, d.Index = "/", "index.html", true
		}
		if d.Meta.Title == "" {
			d.Meta.Title = titleOf(d)
		}
		if d.Meta.Nav == "" {
			d.Meta.Nav = label(slug)
		}
		if d.Meta.Description == "" {
			d.Meta.Description = d.Meta.Subtitle
		}
		docs = append(docs, d)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("%s: %w", s.dir, ErrNoPages)
	}
	sort.SliceStable(docs, func(i, j int) bool {
		if docs[i].Index != docs[j].Index {
			return docs[i].Index
		}
		if docs[i].Meta.Order != docs[j].Meta.Order {
			return docs[i].Meta.Order < docs[j].Meta.Order
		}
		return docs[i].Slug < docs[j].Slug
	})

	return docs, nil
}

func titleOf(d Doc) string {
	for _, line := range strings.Split(string(d.Body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			if t := strings.TrimSpace(line[2:]); t != "" {
				return t
			}
		}
	}
	return d.Slug
}

func label(slug string) string {
	r, size := utf8.DecodeRuneInString(slug)
	if r == utf8.RuneError {
		return slug
	}
	return string(unicode.ToUpper(r)) + slug[size:]
}

func NavFrom(docs []Doc) []NavItem {
	items := make([]NavItem, 0, len(docs))
	for _, d := range docs {
		items = append(items, NavItem{Label: d.Meta.Nav, Path: d.Path})
	}
	return items
}
