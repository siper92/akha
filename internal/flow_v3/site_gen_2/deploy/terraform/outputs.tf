output "public_ip" {
  value = aws_instance.site.public_ip
}

output "url" {
  value = "http://${aws_instance.site.public_ip}"
}

output "instance_id" {
  value = aws_instance.site.id
}

output "bucket" {
  value = aws_s3_bucket.artifact.id
}
