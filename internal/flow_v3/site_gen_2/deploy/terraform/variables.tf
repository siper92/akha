variable "region" {
  type    = string
  default = "eu-central-1"
}

variable "project" {
  type    = string
  default = "site-gen-2"
}

variable "instance_type" {
  type    = string
  default = "t4g.micro"
}

variable "dist_dir" {
  type    = string
  default = "../../dist"
}

variable "ssh_cidr" {
  type    = string
  default = ""
}

variable "key_name" {
  type    = string
  default = ""
}

variable "domain" {
  type    = string
  default = ""
}

variable "zone_name" {
  type    = string
  default = ""
}

variable "www" {
  type    = bool
  default = true
}

variable "acme_email" {
  type    = string
  default = ""
}

variable "eip_allocation_id" {
  type    = string
  default = ""
}

variable "vpc_id" {
  type    = string
  default = ""
}

variable "subnet_id" {
  type    = string
  default = ""
}

variable "assume_role_arn" {
  type    = string
  default = ""
}
