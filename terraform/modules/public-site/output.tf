output "bucket" {
  value = aws_s3_bucket.public.id
}

output "url" {
  value = aws_s3_bucket.public.website_endpoint
}
