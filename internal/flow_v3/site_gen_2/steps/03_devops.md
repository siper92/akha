# step 3 - devops (sub agent 2)

## target

- aws, one `t4g.micro` (graviton 2, arm64)
- ami: amazon linux 2023 arm64 from the ssm parameter `/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64`
- default vpc and a public subnet

## artifact

- the user builds `dist/` before `terraform apply`:
  - `dist/site` - binary, `GOOS=linux GOARCH=arm64 CGO_ENABLED=0`
  - `dist/md/`, `dist/web/`, `dist/config.yaml`
- terraform zips `dist/` with the `archive_file` data source
- the zip is uploaded to a private s3 bucket (`aws_s3_object`, etag from the zip hash so a change redeploys)

## instance

- iam role + instance profile: read only on that one s3 object, plus `AmazonSSMManagedInstanceCore` for shell access without ssh
- security group: 80 from `0.0.0.0/0`, 22 only from `var.ssh_cidr` when set (default none)
- `user_data` (templatefile): download zip, unpack to `/opt/site`, create user `site`, install a systemd unit
- systemd: `WorkingDirectory=/opt/site`, `ExecStart=/opt/site/site serve`, `Environment=SITE_ADDR=:80`, `AmbientCapabilities=CAP_NET_BIND_SERVICE`, `Restart=always`
- `user_data_replace_on_change = true` so a new zip hash replaces the instance

## files

- `deploy/terraform/versions.tf`, `variables.tf`, `main.tf`, `iam.tf`, `outputs.tf`
- `deploy/terraform/user_data.sh.tftpl`
- `deploy/terraform/terraform.tfvars.example`
- `deploy/README.md` - build, apply, destroy, short lists only

## variables

- `region` (`eu-central-1`), `project` (`site-gen-2`), `instance_type` (`t4g.micro`), `dist_dir` (`../../dist`), `ssh_cidr` (empty), `key_name` (empty)

## outputs

- `public_ip`, `url`, `instance_id`, `bucket`
