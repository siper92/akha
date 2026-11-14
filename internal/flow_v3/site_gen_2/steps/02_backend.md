# step 2 - go backend (sub agent 1)

## module

- `site_gen_2/go.mod`, module `sitegen`, go 1.23
- requires: gin, cobra, viper, goldmark
- no go.sum can be generated in this run, the user runs `go mod tidy`

## packages and interfaces

### `internal/config`

- `type Config struct` - Addr, MDDir, Template, StaticDir, IndexPage, SiteName (mapstructure tags = config keys)
- `func Load(v *viper.Viper) (Config, error)` - defaults, env prefix `SITE_`, validation
- sentinel: `ErrInvalidConfig`

### `internal/pages`

- `pages.go` holds the interfaces:
  - `Source` - `List(ctx) ([]Doc, error)` - reads md files
  - `Renderer` - `Render(ctx, Doc) (Page, error)` - md -> html, merged with the template
  - `Store` - `Get(path string) (Page, bool)`, `Nav() []NavItem`
- types: `Doc{Slug, Path string; Body []byte}`, `Page{Slug, Path, Title string; HTML []byte}`, `NavItem{Label, Path string; Active bool}`, `View` (template data, see shared contract)
- impls in their own files:
  - `dirSource` - `NewDirSource(dir, indexPage string) Source`, sorted, index page first
  - `templateRenderer` - `NewRenderer(tmplPath, siteName string, nav []NavItem) (Renderer, error)`, goldmark with GFM, html/template with `.Content` as `template.HTML`
  - `memStore` - `Build(ctx, Source, Renderer) (Store, error)`, renders every page once on boot
- title = first h1 text of the md, fallback slug
- every impl asserted with `var _ I = (*impl)(nil)`
- sentinels: `ErrNoPages`, `ErrTemplate`

### `internal/server`

- `func New(cfg config.Config, store pages.Store) *gin.Engine`
- `gin.New()` + `gin.Recovery()` + a slog request logger
- routes: `/healthz`, `/static/*filepath` from `StaticDir`, one GET per page, NoRoute -> 404
- page responses: `text/html; charset=utf-8`, bytes from the store
- `func Run(ctx, *gin.Engine, addr) error` - `http.Server` with timeouts, graceful shutdown on ctx cancel

### `cmd/`

- `root.go` - root command `site`, `--config` persistent flag, viper init
- `serve.go` - `serve` command, flags `--addr`, `--md-dir` bound to viper, boot order:
  - load config -> source -> renderer -> build store -> server -> run until SIGINT/SIGTERM
- `main.go` - `cmd.Execute()`

## rules

- `context.Context` first argument
- sentinel errors with `fmt.Errorf("%w")`
- no comments, no tests
- `log/slog` for logs
