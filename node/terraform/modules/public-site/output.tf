output "bucket" {
  value = aws_s3_bucket.public.id
}

output "url" {
  value = aws_s3_bucket_website_configuration.public.website_endpoint
}
