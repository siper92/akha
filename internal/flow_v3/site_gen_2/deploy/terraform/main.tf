provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project = var.project
    }
  }

  dynamic "assume_role" {
    for_each = var.assume_role_arn != "" ? [var.assume_role_arn] : []
    content {
      role_arn     = assume_role.value
      session_name = "${var.project}-terraform"
    }
  }
}

locals {
  dist_path = "${path.module}/${var.dist_dir}"
  zip_key   = "site.zip"
  tls_on    = var.domain != ""
  hostnames = local.tls_on ? (var.www ? [var.domain, "www.${var.domain}"] : [var.domain]) : []
  zone_name = var.zone_name != "" ? var.zone_name : var.domain
  vpc_id    = var.vpc_id != "" ? var.vpc_id : data.aws_vpc.default[0].id
  subnet_id = var.subnet_id != "" ? var.subnet_id : sort(data.aws_subnets.default.ids)[0]

  eip_allocation_id = var.eip_allocation_id != "" ? data.aws_eip.existing[0].id : aws_eip.site[0].id
  eip_public_ip     = var.eip_allocation_id != "" ? data.aws_eip.existing[0].public_ip : aws_eip.site[0].public_ip
}

data "aws_ssm_parameter" "ami" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
}

data "aws_vpc" "default" {
  count   = var.vpc_id == "" ? 1 : 0
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }
}

data "archive_file" "dist" {
  type        = "zip"
  source_dir  = local.dist_path
  output_path = "${path.module}/.build/site.zip"
}

resource "random_id" "suffix" {
  byte_length = 4
}

resource "aws_s3_bucket" "artifact" {
  bucket        = "${var.project}-artifact-${random_id.suffix.hex}"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "artifact" {
  bucket                  = aws_s3_bucket.artifact.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "artifact" {
  bucket = aws_s3_bucket.artifact.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_object" "artifact" {
  bucket      = aws_s3_bucket.artifact.id
  key         = local.zip_key
  source      = data.archive_file.dist.output_path
  source_hash = data.archive_file.dist.output_base64sha256
  etag        = data.archive_file.dist.output_md5
}

resource "aws_security_group" "site" {
  name        = "${var.project}-sg"
  description = "site http https and optional ssh"
  vpc_id      = local.vpc_id

  ingress {
    description = "http"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "https"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  dynamic "ingress" {
    for_each = var.ssh_cidr != "" ? [var.ssh_cidr] : []
    content {
      description = "ssh"
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_eip" "site" {
  count  = var.eip_allocation_id == "" ? 1 : 0
  domain = "vpc"

  tags = {
    Name = "${var.project}-eip"
  }

  lifecycle {
    prevent_destroy = false
  }
}

data "aws_eip" "existing" {
  count = var.eip_allocation_id != "" ? 1 : 0
  id    = var.eip_allocation_id
}

resource "aws_instance" "site" {
  ami                         = data.aws_ssm_parameter.ami.value
  instance_type               = var.instance_type
  subnet_id                   = local.subnet_id
  vpc_security_group_ids      = [aws_security_group.site.id]
  iam_instance_profile        = aws_iam_instance_profile.site.name
  key_name                    = var.key_name == "" ? null : var.key_name
  associate_public_ip_address = true

  user_data = templatefile("${path.module}/user_data.sh.tftpl", {
    bucket      = aws_s3_bucket.artifact.id
    key         = aws_s3_object.artifact.key
    region      = var.region
    zip_hash    = data.archive_file.dist.output_base64sha256
    tls_domains = join(",", local.hostnames)
    tls_email   = var.acme_email
  })
  user_data_replace_on_change = true

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }

  root_block_device {
    volume_type = "gp3"
    volume_size = 8
    encrypted   = true
  }

  tags = {
    Name    = "${var.project}-site"
    ZipHash = data.archive_file.dist.output_base64sha256
  }

  depends_on = [aws_s3_object.artifact]
}

resource "aws_eip_association" "site" {
  instance_id   = aws_instance.site.id
  allocation_id = local.eip_allocation_id
}

data "aws_route53_zone" "site" {
  count        = local.tls_on ? 1 : 0
  name         = local.zone_name
  private_zone = false
}

resource "aws_route53_record" "site" {
  for_each        = toset(local.hostnames)
  zone_id         = data.aws_route53_zone.site[0].zone_id
  name            = each.value
  type            = "A"
  ttl             = 300
  records         = [local.eip_public_ip]
  allow_overwrite = true
}
