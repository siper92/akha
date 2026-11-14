# site_gen_2

A small static site. A go server merges markdown pages into one html template on boot and serves them from memory. Terraform deploys it to one aws graviton instance with https.

## layout

- `md/` - one markdown file per page, `world.md` is served at `/`, `home.md` at `/home`
- `web/templates/layout.html` - the page template, the markdown goes into `.Content`
- `web/static/` - css and the logo
- `cmd/`, `internal/` - the go server (gin, cobra, viper, goldmark)
- `config.yaml` - server config, every key can be overridden with a `SITE_` env var
- `deploy/terraform/` - aws infrastructure
- `docs/` - guides, start with `docs/amazon_setup.md`
- `steps/` - the plan the site was built from
- `Justfile` - all commands

## run locally

- `just tidy` - first run only, creates `go.sum`
- `just run` - serves on `http://localhost:8080`
- `just health` - checks `/healthz`

## deploy

One-time setup (aws account, terraform role, vpc, elastic ip, dns, https) lives in `docs/amazon_setup.md`. After that:

- `cp deploy/terraform/terraform.tfvars.example deploy/terraform/terraform.tfvars` and edit it
- `just tf-init`
- `just tf-plan` - builds `dist/` and shows the plan
- `just deploy` - builds `dist/` and applies
- `just url` and `just remote-health`
- redeploy - edit `md/` or `web/`, then `just deploy`
- `just destroy` - removes everything, including the bucket

## useful commands

- `just` - list all recipes
- `just dns` - hostnames and public ip
- `just http-health` - check `/healthz` over http on the ip
- `just remote-health` - check the deployed site
- `just clean` - remove build output
