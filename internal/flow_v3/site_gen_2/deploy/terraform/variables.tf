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
