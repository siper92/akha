package pages

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type templateRenderer struct {
	tmpl     *template.Template
	md       goldmark.Markdown
	siteName string
	nav      []NavItem
}

var _ Renderer = (*templateRenderer)(nil)

func NewRenderer(tmplPath, siteName string, nav []NavItem) (Renderer, error) {
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTemplate, err)
	}
	return &templateRenderer{
		tmpl:     t,
		md:       goldmark.New(goldmark.WithExtensions(extension.GFM)),
		siteName: siteName,
		nav:      nav,
	}, nil
}

func (r *templateRenderer) Render(ctx context.Context, d Doc) (Page, error) {
	if err := ctx.Err(); err != nil {
		return Page{}, err
	}
	var body bytes.Buffer
	if err := r.md.Convert(d.Body, &body); err != nil {
		return Page{}, fmt.Errorf("%w: %s: %v", ErrRender, d.Slug, err)
	}
	title := titleOf(d)
	nav := make([]NavItem, len(r.nav))
	for i, item := range r.nav {
		item.Active = item.Path == d.Path
		nav[i] = item
	}
	view := View{
		Title:    title,
		Slug:     d.Slug,
		Path:     d.Path,
		Content:  template.HTML(body.String()),
		Nav:      nav,
		SiteName: r.siteName,
		Year:     time.Now().Year(),
	}
	var out bytes.Buffer
	if err := r.tmpl.Execute(&out, view); err != nil {
		return Page{}, fmt.Errorf("%w: %s: %v", ErrRender, d.Slug, err)
	}
	return Page{Slug: d.Slug, Path: d.Path, Title: title, HTML: out.Bytes()}, nil
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

func NavFrom(docs []Doc) []NavItem {
	items := make([]NavItem, 0, len(docs))
	for _, d := range docs {
		items = append(items, NavItem{Label: label(d.Slug), Path: d.Path})
	}
	return items
}

func label(slug string) string {
	r, size := utf8.DecodeRuneInString(slug)
	if r == utf8.RuneError {
		return slug
	}
	return string(unicode.ToUpper(r)) + slug[size:]
}
