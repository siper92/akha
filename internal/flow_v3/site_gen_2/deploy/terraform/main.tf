provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project = var.project
    }
  }
}

locals {
  dist_path = "${path.module}/${var.dist_dir}"
  zip_key   = "site.zip"
}

data "aws_ssm_parameter" "ami" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
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
  description = "site http and optional ssh"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description = "http"
    from_port   = 80
    to_port     = 80
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

resource "aws_instance" "site" {
  ami                         = data.aws_ssm_parameter.ami.value
  instance_type               = var.instance_type
  subnet_id                   = sort(data.aws_subnets.default.ids)[0]
  vpc_security_group_ids      = [aws_security_group.site.id]
  iam_instance_profile        = aws_iam_instance_profile.site.name
  key_name                    = var.key_name == "" ? null : var.key_name
  associate_public_ip_address = true

  user_data = templatefile("${path.module}/user_data.sh.tftpl", {
    bucket   = aws_s3_bucket.artifact.id
    key      = aws_s3_object.artifact.key
    region   = var.region
    zip_hash = data.archive_file.dist.output_base64sha256
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
    Name     = "${var.project}-site"
    ZipHash  = data.archive_file.dist.output_base64sha256
  }

  depends_on = [aws_s3_object.artifact]
}
