locals {
  authorizer_lambda_name   = "${var.prefix}-authorizer-${var.suffix}"
  authorizer_lambda_s3_key = "${var.functions_s3_path}/authorizer.zip"
}

resource "aws_lambda_function" "authorizer" {
  count     = var.authorizer == null ? 0 : 1
  role      = aws_iam_role.authorizer[0].arn
  s3_bucket = var.functions_bucket
  s3_key    = local.authorizer_lambda_s3_key

  function_name = local.authorizer_lambda_name
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  environment {
    variables = var.authorizer.env
  }
}

resource "aws_cloudwatch_log_group" "authorizer_log_group" {
  count             = var.authorizer == null ? 0 : 1
  name              = "/aws/lambda/${local.authorizer_lambda_name}"
  retention_in_days = 14
}
