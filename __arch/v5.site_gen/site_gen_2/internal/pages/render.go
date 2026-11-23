package pages

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type RenderOptions struct {
	Template string
	Texts    Texts
	BaseURL  string
	Assets   string
	Lang     string
	Nav      []NavItem
}

type templateRenderer struct {
	tmpl *template.Template
	md   goldmark.Markdown
	opts RenderOptions
	year int
}

var _ Renderer = (*templateRenderer)(nil)

func NewRenderer(opts RenderOptions) (Renderer, error) {
	t, err := template.New(filepath.Base(opts.Template)).
		Funcs(template.FuncMap{"hash": Hash}).
		ParseFiles(opts.Template)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTemplate, err)
	}
	return &templateRenderer{
		tmpl: t,
		md:   goldmark.New(goldmark.WithExtensions(extension.GFM)),
		opts: opts,
		year: time.Now().Year(),
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

	nav := markActive(r.opts.Nav, d.Path)

	img := r.assetURL(d.Meta.Image)
	alt := d.Meta.ImageAlt
	if alt == "" {
		alt = d.Meta.Title
	}
	ogType := "article"
	if d.Index {
		ogType = "website"
	}

	view := View{
		Lang:        r.opts.Lang,
		SiteName:    r.opts.Texts.SiteName,
		Assets:      r.opts.Assets,
		Title:       d.Meta.Title,
		Subtitle:    d.Meta.Subtitle,
		Description: d.Meta.Description,
		Keywords:    strings.Join(d.Meta.Keywords, ", "),
		Author:      d.Meta.Author,
		Canonical:   r.absURL(d.Path),
		Image:       img,
		ImageAbs:    r.absURL(img),
		ImageAlt:    alt,
		OGType:      ogType,
		Slug:        d.Slug,
		Path:        d.Path,
		Content:     template.HTML(body.String()),
		Source:      string(d.Body),
		Texts:       r.opts.Texts,
		Nav:         nav,
		Year:        r.year,
	}
	var out bytes.Buffer
	if err := r.tmpl.Execute(&out, view); err != nil {
		return Page{}, fmt.Errorf("%w: %s: %v", ErrRender, d.Slug, err)
	}
	return Page{Slug: d.Slug, Path: d.Path, File: d.File, Meta: d.Meta, HTML: out.Bytes()}, nil
}

func markActive(items []NavItem, path string) []NavItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]NavItem, len(items))
	for i, it := range items {
		it.Children = markActive(it.Children, path)
		it.Active = it.Path != "" && it.Path == path
		for _, c := range it.Children {
			if c.Active {
				it.Active = true
			}
		}
		out[i] = it
	}
	return out
}

func (r *templateRenderer) assetURL(p string) string {
	if p == "" || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p
	}
	return r.opts.Assets + "/" + p
}

func (r *templateRenderer) absURL(p string) string {
	if p == "" || r.opts.BaseURL == "" {
		return ""
	}
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p
	}
	return r.opts.BaseURL + p
}
