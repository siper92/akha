# step 5 - https and amazon setup

## target

- the go server terminates tls itself with lets encrypt
- a static ip (elastic ip) and route 53 dns for the domain
- a one-time amazon setup guide, the readme only holds run and deploy commands

## server

- `golang.org/x/crypto/acme/autocert`, http-01 challenge on port 80
- env:
  - `SITE_ADDR=:80` - http, acme challenges, redirect to https, `/healthz` still 200 on http
  - `SITE_HTTPS_ADDR=:443`
  - `SITE_TLS_DOMAINS="example.com,www.example.com"` - empty means plain http only
  - `SITE_TLS_EMAIL`
  - `SITE_TLS_CACHE_DIR=/var/lib/site/certs`
- HSTS header when tls is on
- locally `just run` stays plain http on `:8080`

## terraform

- `aws_eip` unless `eip_allocation_id` is set, plus `aws_eip_association` to the instance
- route 53 A records for the domain and `www.<domain>` pointing at the eip
- security group ingress 80, 443 and optional 22
- optional `assume_role_arn`, `vpc_id` and `subnet_id`
- the subnet must be public with a route to an internet gateway

## cert cache

- the cache dir is synced to the artifact s3 bucket under `certs/`
- at boot, and hourly by the systemd timer `site-certs-sync.timer`
- a replaced instance reuses the certs, so the lets encrypt limit of 5 duplicate certs per week is not hit

## docs

- `docs/amazon_setup.md` - account, terraform role, vpc, elastic ip, dns, https, ec2, shell access, first deploy checklist
- `README.md` - short, run and deploy only
- `deploy/README.md` - build, apply, destroy

## variables

- existing: `region` (`eu-central-1`), `project` (`site-gen-2`), `instance_type` (`t4g.micro`), `dist_dir`, `ssh_cidr`, `key_name`
- new:
  - `domain` - empty keeps plain http
  - `zone_name` - route 53 hosted zone, defaults to `domain`
  - `www` - bool, default true, adds `www.<domain>`
  - `acme_email`
  - `eip_allocation_id` - reuse an existing elastic ip, empty means terraform creates one
  - `vpc_id` - empty means the default vpc
  - `subnet_id` - empty means the first subnet of the vpc
  - `assume_role_arn` - role terraform assumes, empty means caller credentials

## outputs

- `public_ip` (the eip), `url` (`https://domain` or `http://ip`), `instance_id`, `bucket`, `eip_allocation_id`, `hostnames`

## just recipes

- `tidy`, `run`, `health`, `dist`, `clean`
- `tf-init`, `tf-fmt`, `tf-validate`, `tf-plan`, `deploy`, `destroy`
- `url`, `id`, `ssh`, `ssm`, `remote-health`
- `http-health` - `curl http://<eip>/healthz`
- `dns` - prints hostnames and `public_ip`
