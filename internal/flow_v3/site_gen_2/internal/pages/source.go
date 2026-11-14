package pages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

		body, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}

		slug := strings.TrimSuffix(e.Name(), ".md")
		path := "/" + slug
		if slug == s.indexPage {
			path = "/"
		}

		docs = append(docs, Doc{Slug: slug, Path: path, Body: body})
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("%s: %w", s.dir, ErrNoPages)
	}
	sort.Slice(docs, func(i, j int) bool {
		ii, ji := docs[i].Slug == s.indexPage, docs[j].Slug == s.indexPage
		if ii != ji {
			return ii
		}
		return docs[i].Slug < docs[j].Slug
	})

	return docs, nil
}
