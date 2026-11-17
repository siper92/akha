# site_gen_2

A small static site generator. One go command merges the markdown pages in `web/md` into the html template and writes a plain static site to `deploy/static`, ready for nginx.

## layout

- `web/md/` - one markdown file per page with yaml front matter (title, description, nav, order, image, keywords)
  - `world.md` is written to `index.html` (`/`), every other page to `<slug>/index.html` (`/<slug>/`)
  - `<page>_outline.md` files are story outlines, they are not built
- `web/templates/layout.html` - the page template, meta and og tags come from the front matter
- `web/static/` - css, js and images, copied to `deploy/static/assets/` (`.md` files are skipped)
- `web/static/design.md` - design and style doc, colors, themes, components, image specs
- `main.go`, `internal/` - the generator (cobra, viper, goldmark, yaml)
- `config.yaml` - build config, every key can be overridden with a `SITE_` env var
- `deploy/terraform/` - aws infrastructure
- `Justfile` - all commands

## build

- `just site` - builds `deploy/static`
- `just site https://example.com` - also writes canonical urls, `og:url`, `og:image` and `sitemap.xml`
- `just preview` - builds and serves `deploy/static` on `http://localhost:8080`

## nginx

```nginx
server {
    root /usr/share/nginx/html;
    index index.html;
    location / {
        try_files $uri $uri/ =404;
    }
}
```

## deploy

- `just dist` - copies `deploy/static` into `dist/`
- `just deploy` - builds `dist/` and applies terraform
