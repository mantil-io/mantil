output "functions" {
  value = [for k, v in aws_lambda_function.functions :
    {
      name : k
      arn : v.arn
      invoke_arn : v.invoke_arn
    }
  ]
}

output "static_websites" {
  value = [for k, v in local.static_websites :
    {
      name : v.name
      bucket : aws_s3_bucket.static_websites[k].id
      url : aws_s3_bucket.static_websites[k].website_endpoint
    }
  ]
}
