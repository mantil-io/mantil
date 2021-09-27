output "functions" {
  value = [for k, v in aws_lambda_function.functions :
    {
      name : k
      arn : v.arn
      invoke_arn : v.invoke_arn
    }
  ]
}
output "role_arn" {
  value = aws_iam_role.lambda.arn
}
