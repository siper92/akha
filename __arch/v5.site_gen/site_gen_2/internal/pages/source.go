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

	docs, err := s.readDir(ctx, s.dir, "")
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", s.dir, err)
	}
	for _, e := range entries {
		if !e.IsDir() || skipName(e.Name()) {
			continue
		}
		sub, err := s.readDir(ctx, filepath.Join(s.dir, e.Name()), e.Name())
		if err != nil {
			return nil, err
		}
		docs = append(docs, sub...)
	}

	seen := make(map[string]string, len(docs))
	for _, d := range docs {
		if g, ok := seen[d.Slug]; ok {
			return nil, fmt.Errorf("%w: %s in %q and %q", ErrDupSlug, d.Slug, g, d.Group)
		}
		seen[d.Slug] = d.Group
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

func (s *dirSource) readDir(ctx context.Context, dir, group string) ([]Doc, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	var docs []Doc
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		if strings.HasSuffix(slug, outlineSuffix) || skipName(slug) {
			continue
		}

		file := filepath.Join(dir, e.Name())
		raw, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		meta, body, err := splitFrontMatter(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		if meta.Draft {
			continue
		}

		d := Doc{Slug: slug, Group: group, Path: "/" + slug + "/", File: slug + "/index.html", Meta: meta, Body: body}
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
	return docs, nil
}

func skipName(name string) bool {
	return strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".")
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

type navEntry struct {
	index bool
	order int
	key   string
	item  NavItem
}

func NavFrom(docs []Doc) []NavItem {
	var entries []navEntry
	groups := make(map[string][]Doc)
	var names []string
	for _, d := range docs {
		if d.Group == "" {
			entries = append(entries, navEntry{
				index: d.Index,
				order: d.Meta.Order,
				key:   d.Slug,
				item:  NavItem{Label: d.Meta.Nav, Path: d.Path},
			})
			continue
		}
		if _, ok := groups[d.Group]; !ok {
			names = append(names, d.Group)
		}
		groups[d.Group] = append(groups[d.Group], d)
	}

	for _, name := range names {
		members := groups[name]
		e := navEntry{order: members[0].Meta.Order, key: name}
		var main []NavItem
		var rest []NavItem
		for _, d := range members {
			if d.Index {
				e.index = true
			}
			it := NavItem{Label: d.Meta.Nav, Path: d.Path}
			if d.Slug == name {
				e.order = d.Meta.Order
				main = append(main, it)
				continue
			}
			rest = append(rest, it)
		}
		e.item = NavItem{Label: label(name), Children: append(main, rest...)}
		entries = append(entries, e)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].index != entries[j].index {
			return entries[i].index
		}
		if entries[i].order != entries[j].order {
			return entries[i].order < entries[j].order
		}
		return entries[i].key < entries[j].key
	})

	items := make([]NavItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, e.item)
	}
	return items
}
