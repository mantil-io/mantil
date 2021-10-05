resource "aws_s3_bucket" "static_websites" {
  for_each = local.static_websites
  bucket_prefix = "mantil-public-${var.project_name}-${each.value.name}-"
  acl    = "public-read"
  force_destroy = true

  website {
    index_document = "index.html"
    error_document = "index.html"
  }
}

resource "aws_s3_bucket_policy" "public_read" {
  for_each = aws_s3_bucket.static_websites
  bucket = each.value.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource = [
          each.value.arn,
          "${each.value.arn}/*",
        ]
      },
    ]
  })
}
