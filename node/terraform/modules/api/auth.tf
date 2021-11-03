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

resource "aws_lambda_permission" "authorizer_http_api_gateway_invoke" {
  count         = var.authorizer == null ? 0 : 1
  function_name = aws_lambda_function.authorizer[0].arn
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http.execution_arn}/*/*"
}

resource "aws_lambda_permission" "authorizer_ws_api_gateway_invoke" {
  count         = var.authorizer == null ? 0 : 1
  function_name = aws_lambda_function.authorizer[0].arn
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.ws.execution_arn}/*/*"
}

resource "aws_apigatewayv2_authorizer" "ws" {
  count            = var.authorizer == null ? 0 : 1
  api_id           = aws_apigatewayv2_api.ws.id
  authorizer_type  = "REQUEST"
  authorizer_uri   = aws_lambda_function.authorizer[0].invoke_arn
  identity_sources = ["route.request.header.${var.authorizer.authorization_header}"]
  name             = "${var.prefix}-ws-authorizer-${var.suffix}"
}

resource "aws_apigatewayv2_authorizer" "http" {
  count                             = var.authorizer == null ? 0 : 1
  api_id                            = aws_apigatewayv2_api.http.id
  authorizer_type                   = "REQUEST"
  authorizer_uri                    = aws_lambda_function.authorizer[0].invoke_arn
  identity_sources                  = ["$request.header.${var.authorizer.authorization_header}"]
  authorizer_payload_format_version = "1.0"
  name                              = "${var.prefix}-http-authorizer-${var.suffix}"
  authorizer_result_ttl_in_seconds  = 0
}
