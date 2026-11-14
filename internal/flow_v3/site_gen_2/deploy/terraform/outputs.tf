output "public_ip" {
  value = local.eip_public_ip
}

output "url" {
  value = local.tls_on ? "https://${var.domain}" : "http://${local.eip_public_ip}"
}

output "eip_allocation_id" {
  value = local.eip_allocation_id
}

output "hostnames" {
  value = local.hostnames
}

output "instance_id" {
  value = aws_instance.site.id
}

output "bucket" {
  value = aws_s3_bucket.artifact.id
}
