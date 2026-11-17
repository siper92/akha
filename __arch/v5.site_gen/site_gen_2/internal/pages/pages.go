package pages

import (
	"context"
	"errors"
	"html/template"
)

var (
	ErrNoPages  = errors.New("no pages")
	ErrTemplate = errors.New("template error")
	ErrRender   = errors.New("render error")
)

type Doc struct {
	Slug string
	Path string
	Body []byte
}

type Page struct {
	Slug  string
	Path  string
	Title string
	HTML  []byte
}

type NavItem struct {
	Label  string
	Path   string
	Active bool
}

type View struct {
	Title    string
	Slug     string
	Path     string
	Content  template.HTML
	Nav      []NavItem
	SiteName string
	Year     int
}

type Source interface {
	List(ctx context.Context) ([]Doc, error)
}

type Renderer interface {
	Render(ctx context.Context, d Doc) (Page, error)
}

type Store interface {
	Get(path string) (Page, bool)
	Pages() []Page
	Nav() []NavItem
}
