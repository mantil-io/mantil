output "functions" {
  value = [for k, v in aws_lambda_function.functions :
    {
      name : k
      arn : v.arn
      invoke_arn : v.invoke_arn
    }
  ]
}

output "public" {
  value = {
    bucket : aws_s3_bucket.public.id
    url : aws_s3_bucket.public.website_endpoint
  }
}
