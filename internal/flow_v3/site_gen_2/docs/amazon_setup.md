# amazon setup

One-time setup guide for deploying site_gen_2 to aws with a domain, a static ip and https.

## overview

Done once by hand:

- aws account, cli and tools
- a route 53 hosted zone for your domain
- an iam role terraform assumes
- optional: own vpc and subnet
- optional: a pre-allocated elastic ip

Done by terraform:

- s3 bucket, iam role and instance profile for the instance
- security group (80, 443, optional 22)
- ec2 instance, elastic ip and its association
- route 53 A records for the domain and www

Request path:

```
dns (route 53)
   |
   v
elastic ip
   |
   v
ec2 in a public subnet of a vpc
   |
   v
go server :80 (acme, redirect, healthz) and :443 (site)
```

## prerequisites

- aws account with billing enabled
- aws cli v2
- session-manager-plugin for the aws cli
- terraform >= 1.6
- go toolchain
- a domain whose hosted zone is in route 53

Check the zone:

```
aws route53 list-hosted-zones
```

- the NS records at your registrar must match the NS set of the hosted zone
- without a matching NS set, lets encrypt cannot reach your server

## terraform role

Terraform assumes a dedicated iam role, so day to day credentials stay small.

Trust policy, saved as `trust.json`. Replace `111122223333` with your account id and the user with your iam user or sso principal:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "AWS": "arn:aws:iam::111122223333:user/your-user" },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

Permissions policy, saved as `permissions.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["ec2:*", "s3:*"],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "iam:*",
      "Resource": [
        "arn:aws:iam::111122223333:role/site-gen-2-*",
        "arn:aws:iam::111122223333:instance-profile/site-gen-2-*",
        "arn:aws:iam::111122223333:policy/site-gen-2-*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "arn:aws:iam::111122223333:role/site-gen-2-*"
    },
    {
      "Effect": "Allow",
      "Action": ["ssm:GetParameter", "ssm:GetParameters"],
      "Resource": "arn:aws:ssm:*::parameter/aws/service/ami-amazon-linux-latest/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets",
        "route53:GetHostedZone",
        "route53:ListHostedZones",
        "route53:ListResourceRecordSets",
        "route53:GetChange"
      ],
      "Resource": "*"
    }
  ]
}
```

- narrow `ec2:*` and `s3:*` further if you want, they are kept wide here for brevity
- `iam:*` is limited to resources named `site-gen-2-*`

Create the role:

```
aws iam create-role --role-name site-terraform --assume-role-policy-document file://trust.json
aws iam put-role-policy --role-name site-terraform --policy-name site-terraform --policy-document file://permissions.json
```

Add a profile to `~/.aws/config`:

```
[profile site-terraform]
role_arn = arn:aws:iam::111122223333:role/site-terraform
source_profile = default
region = eu-central-1
```

Two ways to use it:

- set `assume_role_arn` in `terraform.tfvars`, terraform assumes the role itself
- or run with the profile: `AWS_PROFILE=site-terraform just deploy`

Verify:

```
aws sts get-caller-identity
AWS_PROFILE=site-terraform aws sts get-caller-identity
```

- the second call must show an `assumed-role/site-terraform` arn

## vpc

The instance needs a public subnet. Pick one option.

### option a - default vpc

- leave `vpc_id` and `subnet_id` empty
- check that it exists:

```
aws ec2 describe-vpcs --filters Name=isDefault,Values=true
```

- if it was deleted, recreate it:

```
aws ec2 create-default-vpc
```

### option b - own vpc

```
aws ec2 create-vpc --cidr-block 10.0.0.0/16 --tag-specifications 'ResourceType=vpc,Tags=[{Key=Name,Value=site-gen-2}]'
aws ec2 modify-vpc-attribute --vpc-id <vpc-id> --enable-dns-support
aws ec2 modify-vpc-attribute --vpc-id <vpc-id> --enable-dns-hostnames
aws ec2 create-subnet --vpc-id <vpc-id> --cidr-block 10.0.1.0/24 --availability-zone eu-central-1a
aws ec2 modify-subnet-attribute --subnet-id <subnet-id> --map-public-ip-on-launch
aws ec2 create-internet-gateway
aws ec2 attach-internet-gateway --internet-gateway-id <igw-id> --vpc-id <vpc-id>
aws ec2 create-route-table --vpc-id <vpc-id>
aws ec2 create-route --route-table-id <rtb-id> --destination-cidr-block 0.0.0.0/0 --gateway-id <igw-id>
aws ec2 associate-route-table --route-table-id <rtb-id> --subnet-id <subnet-id>
```

Then set in `terraform.tfvars`:

```
vpc_id    = "vpc-..."
subnet_id = "subnet-..."
```

### why the subnet must be public

- the elastic ip only works on an instance in a subnet with a route to an internet gateway
- lets encrypt must reach port 80 of the instance (http-01)
- the instance downloads the zip from s3 and the ami tools over the internet
- an empty `subnet_id` means the first subnet of the vpc, it must be public

## static ip with an elastic ip

A fixed ip keeps dns records valid across instance replacements.

Terraform managed (default):

- leave `eip_allocation_id` empty
- terraform creates an `aws_eip` and associates it
- the ip is lost on `just destroy`

Pre-allocated:

```
aws ec2 allocate-address --domain vpc --tag-specifications 'ResourceType=elastic-ip,Tags=[{Key=Name,Value=site-gen-2}]'
```

- copy the `AllocationId` into `eip_allocation_id` in `terraform.tfvars`
- the ip survives `just destroy`, terraform only associates it

Bring your own ip range (BYOIP):

```
aws ec2 describe-public-ipv4-pools
aws ec2 allocate-address --domain vpc --public-ipv4-pool <ipv4pool-ec2-...>
```

- takes an address from an ip range you own and have provisioned in aws
- use the returned allocation id as `eip_allocation_id`

Cost:

- a public ipv4 address is billed hourly
- an unattached elastic ip is billed too, release it when you no longer need it:
  `aws ec2 release-address --allocation-id <eipalloc-...>`

## dns in route 53

- set `domain` in `terraform.tfvars`
- terraform writes A records for the domain and, with `www = true`, for `www.<domain>`
- both point at the elastic ip
- `zone_name` defaults to `domain`
- set `zone_name` when the zone has another name, for example a parent zone, or was created by hand
- a zone in another aws account is not reachable by this role, create the records there by hand

Check:

```
dig +short example.com
just dns
```

- both must print the `public_ip`
- dns changes can take a few minutes to spread

## https

How it works:

- the go server gets certificates from lets encrypt with `autocert`
- challenge type is http-01 on port 80
- certificates are cached in `/var/lib/site/certs`
- the cache is synced to the artifact s3 bucket under `certs/` at boot and hourly by `site-certs-sync.timer`
- a replaced instance restores the certs from s3 and does not hit the lets encrypt limit of 5 duplicate certs per week
- renewal is automatic, about 30 days before expiry
- port 80 redirects to https, `/healthz` still answers 200 on http
- the server sends an HSTS header when tls is on

Steps:

- set `domain` and `acme_email` in `terraform.tfvars`
- `just deploy`
- wait for dns to point at the elastic ip
- the first https request triggers issuance, it can take a few seconds
- `just remote-health`
- `curl -I https://example.com` shows `strict-transport-security`

Troubleshooting:

- dns does not point at the elastic ip yet - check `dig +short example.com`
- port 80 is closed - the security group must allow 80 and 443, check with `just http-health`
- rate limits - wait, or reuse certs through the s3 cache, do not destroy and redeploy in a loop
- read the server log over ssm:

```
just ssm
sudo journalctl -u site -n 100
```

## ec2

Request flow on the instance:

```
internet
   |
   | http :80, https :443
   v
security group  (80 and 443 open, 22 closed by default)
   |
   v
ec2 t4g.micro  (amazon linux 2023, arm64, public subnet, elastic ip)
   |  systemd: site.service -> /opt/site/site serve
   |
   | s3:GetObject (site.zip), s3 certs/ sync (instance role)
   v
s3 bucket (private)  <-  site.zip  <-  terraform  <-  dist/
```

Parts:

- s3 bucket
  - private, all public access blocked, AES256 encryption
  - holds `site.zip`, the zipped `dist/` folder, and `certs/`
  - a new zip hash replaces the instance, so a deploy is a new instance
- iam
  - the instance role can read the zip object and read and write `certs/`
  - `AmazonSSMManagedInstanceCore` is attached for shell access without ssh
- security group
  - inbound 80 and 443 from anywhere
  - inbound 22 only when `ssh_cidr` is set, closed by default
  - outbound all
- elastic ip
  - created by terraform or reused through `eip_allocation_id`
  - associated with the instance by `aws_eip_association`
- ec2 instance
  - `t4g.micro`, graviton 2, amazon linux 2023 arm64 ami from ssm
  - IMDSv2 required, 8GB encrypted gp3 root volume
  - user data downloads the zip, unpacks it to `/opt/site`, restores certs and starts `site.service`
  - the service runs as the `site` user on ports 80 and 443 with `CAP_NET_BIND_SERVICE`
  - `site-certs-sync.timer` syncs the cert cache to s3 every hour
  - boot log: `/var/log/site-bootstrap.log`

## shell access

Port 22 is closed by default. There are four ways to reach the instance, from most to least recommended.

### 1 - ssm session manager (default, no ssh)

- no open port, no key, no public ssh endpoint
- access is controlled by iam and every session is logged in cloudtrail
- needs the aws cli and the session manager plugin on your machine
- `just ssm`

### 2 - ssh over ssm (real ssh, port 22 still closed)

Use this for `scp`, `rsync`, port forwarding or tools that need real ssh.

- set `key_name` to an existing ec2 key pair, keep `ssh_cidr` empty
- add to `~/.ssh/config`:

```
Host i-*
    User ec2-user
    IdentityFile ~/.ssh/<your-key>.pem
    ProxyCommand aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters portNumber=%p
```

- connect with `just ssh` or `ssh i-0123456789abcdef0`
- the tunnel goes through ssm, the security group needs no inbound 22 rule

### 3 - ssh from a fixed ip

- set `ssh_cidr` to your public ip as `/32`, for example `203.0.113.10/32`
- set `key_name` to an existing ec2 key pair
- `just deploy`, then `ssh -i ~/.ssh/<your-key>.pem ec2-user@<public_ip>`
- never use `0.0.0.0/0`, scanners find an open 22 within minutes
- remove `ssh_cidr` and apply again when you are done

### 4 - ec2 instance connect

- amazon linux 2023 ships the instance connect agent
- pushes a short lived (60s) public key through iam, no key pair needed
- `aws ec2-instance-connect send-ssh-public-key --instance-id <id> --instance-os-user ec2-user --ssh-public-key file://~/.ssh/id_ed25519.pub`
- then ssh within 60 seconds, directly or over ssm as in option 2
- direct ssh needs inbound 22 from your ip or from the `com.amazonaws.<region>.ec2-instance-connect` prefix list

### hardening

- keep 22 closed and prefer options 1 and 2
- use ed25519 keys, never share a key between people
- give people iam permissions for `ssm:StartSession` on tagged instances only
- turn on session logging to s3 or cloudwatch in the ssm settings
- the instance is replaced on every deploy, so treat it as disposable and never fix things by hand on it

## first deploy checklist

- install aws cli v2, session-manager-plugin, terraform, go
- confirm the hosted zone and the registrar NS records
- create the terraform role and the `site-terraform` profile
- check the default vpc, or create your own and note `vpc_id` and `subnet_id`
- optional: allocate an elastic ip and note the allocation id
- `cp deploy/terraform/terraform.tfvars.example deploy/terraform/terraform.tfvars` and fill in `domain`, `acme_email` and the optional values
- `just tf-init`
- `just tf-plan`
- `just deploy`
- `just dns` and `dig +short <domain>`
- `just http-health`
- `just remote-health`
- `curl -I https://<domain>`
