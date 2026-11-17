# deploy

Terraform for one t4g.micro running the site on ports 80 and 443 behind an elastic ip.

One-time setup (role, vpc, elastic ip, dns, https) is in `../docs/amazon_setup.md`.

## prerequisites

- terraform >= 1.6
- go toolchain
- aws credentials with rights for ec2, s3, iam, ssm, route 53

## build dist

Run from site_gen_2:

```
rm -rf dist && mkdir -p dist
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/site .
cp -r md web config.yaml dist/
```

## apply

```
cd deploy/terraform
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform plan
terraform apply
```

## check

```
curl "$(terraform output -raw url)/healthz"
curl -I "$(terraform output -raw url)"
```

- expect 200 ok
- with `domain` set the url is `https://<domain>`, otherwise `http://<ip>`
- bootstrap takes about a minute after apply
- the first https request issues the certificate, dns must point at the elastic ip
- log on the instance: /var/log/site-bootstrap.log

## redeploy

- rebuild dist, then run `terraform apply`
- a changed zip hash replaces the instance
- the elastic ip and the cached certs (s3 `certs/`) are reused

## destroy

```
terraform destroy
```

- the bucket is emptied automatically (force_destroy)
- a terraform-created elastic ip is released, a pre-allocated one (`eip_allocation_id`) is kept
