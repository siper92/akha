# deploy

Terraform for one t4g.micro running the site behind port 80.

## prerequisites

- terraform >= 1.6
- go toolchain
- aws credentials with rights for ec2, s3, iam, ssm
- optional: session-manager-plugin for the aws cli

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
```

- expect 200 ok
- bootstrap takes about a minute after apply
- log on the instance: /var/log/site-bootstrap.log

## ssm session

```
aws ssm start-session --target "$(terraform output -raw instance_id)" --region eu-central-1
```

## redeploy

- rebuild dist, then run `terraform apply`
- a changed zip hash replaces the instance

## destroy

```
terraform destroy
```

- the bucket is emptied automatically (force_destroy)
