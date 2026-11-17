# site_gen_2 - plan

## goal

- a small static site served by a go server
- md files are merged into one html template on boot, pages live in memory
- two pages: `world` at `/`, `home` at `/home`
- deploy to aws on a graviton 2 micro instance with terraform

## libraries and languages

- go 1.23+ (own module, `site_gen_2/go.mod`, module path `sitegen`)
- github.com/gin-gonic/gin - http server and routing
- github.com/spf13/cobra - cli, `serve` command
- github.com/spf13/viper - config file, env, flags
- github.com/yuin/goldmark - md to html (with GFM extension)
- everything else is go standard library (`html/template`, `embed` not used, `os`, `path/filepath`, `log/slog`, `context`)
- html5 + plain css, no js, no css framework
- fonts from google fonts with a system fallback stack
- terraform >= 1.6, aws provider ~> 5.0, archive provider

## data sources

- `md/*.md` - one file per page, file name is the page slug
- `web/templates/layout.html` - the single page template
- `web/static/` - css and the svg logo
- `config.yaml` - server config

## directory layout

- `steps/` - these plan files
- `md/` - `world.md`, `home.md`
- `web/templates/layout.html`
- `web/static/css/style.css`
- `web/static/img/logo.svg`
- `main.go`, `cmd/`, `internal/config/`, `internal/pages/`, `internal/server/`
- `config.yaml`
- `deploy/terraform/` - terraform and the instance boot script

## order of work

- step 1 - design and style (sub agent 0) - template, css, logo, lorem ipsum md files
- step 2 - go backend (sub agent 1) - runs in parallel with step 1, codes against the contract below
- step 3 - devops (sub agent 2) - runs in parallel, codes against the paths and config below
- step 4 - content (sub agent 4) - after step 1, replaces lorem ipsum with the story

## shared contract - template data

the template is executed with one value of type `pages.View`

- `.Title` string - page title (first `# h1` of the md, fallback: slug)
- `.Slug` string - `world`, `home`
- `.Path` string - `/`, `/home`
- `.Content` template.HTML - the rendered md
- `.Nav` []NavItem - `{Label, Path, Active bool}` in nav order: world (`/`), home (`/home`)
- `.SiteName` string - from config, default `Akha Collective`
- `.Year` int - for the footer

static files are served at `/static/...` from `web/static/`

## shared contract - md format

- line 1: `# Title`
- line 2: `## Subtitle` (styled as the hero subtitle)
- then paragraphs, allowed: `**bold**`, `*em*`, `> quote`, `-` lists, `---`, `[links](/home)`, inline `code`
- no raw html, no images, no front matter

## shared contract - config (viper)

- keys: `addr` (`:8080`), `md_dir` (`md`), `template` (`web/templates/layout.html`), `static_dir` (`web/static`), `index_page` (`world`), `site_name` (`Akha Collective`)
- file: `config.yaml` in the working dir, or `--config path`
- env: prefix `SITE_`, e.g. `SITE_ADDR=:80`
- flags on `serve`: `--addr`, `--md-dir`

## routing rules

- every `md/<slug>.md` becomes `/<slug>`
- the page named by `index_page` is served at `/` instead of `/<slug>`
- unknown path -> 404 with a plain text body
- `GET /healthz` -> 200 `ok`
