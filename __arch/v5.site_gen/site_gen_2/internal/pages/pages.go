package pages

import (
	"context"
	"errors"
	"html/template"
)

var (
	ErrNoPages     = errors.New("no pages")
	ErrTemplate    = errors.New("template error")
	ErrRender      = errors.New("render error")
	ErrFrontMatter = errors.New("front matter error")
	ErrUnsafeOut   = errors.New("unsafe output dir")
	ErrTexts       = errors.New("texts error")
)

type MenuEntry struct {
	Label string   `yaml:"label"`
	Page  string   `yaml:"page"`
	Pages []string `yaml:"pages"`
}

type Texts struct {
	SiteName string      `yaml:"site_name"`
	Slogan   string      `yaml:"slogan"`
	Menu     []MenuEntry `yaml:"menu"`
}

type Meta struct {
	Title       string   `yaml:"title"`
	Subtitle    string   `yaml:"subtitle"`
	Description string   `yaml:"description"`
	Nav         string   `yaml:"nav"`
	Order       int      `yaml:"order"`
	Image       string   `yaml:"image"`
	ImageAlt    string   `yaml:"image_alt"`
	Keywords    []string `yaml:"keywords"`
	Author      string   `yaml:"author"`
	Draft       bool     `yaml:"draft"`
}

type Doc struct {
	Slug  string
	Path  string
	File  string
	Index bool
	Meta  Meta
	Body  []byte
}

type Page struct {
	Slug string
	Path string
	File string
	Meta Meta
	HTML []byte
}

type NavItem struct {
	Label    string
	Path     string
	Active   bool
	Children []NavItem
}

type View struct {
	Lang        string
	SiteName    string
	Assets      string
	Title       string
	Subtitle    string
	Description string
	Keywords    string
	Author      string
	Canonical   string
	Image       string
	ImageAbs    string
	ImageAlt    string
	OGType      string
	Slug        string
	Path        string
	Content     template.HTML
	Source      string
	Texts       Texts
	Nav         []NavItem
	Year        int
}

type Source interface {
	List(ctx context.Context) ([]Doc, error)
}

type Renderer interface {
	Render(ctx context.Context, d Doc) (Page, error)
}

type Writer interface {
	Reset(ctx context.Context) error
	Write(ctx context.Context, pages []Page) error
}

type Assets interface {
	Copy(ctx context.Context) error
}
