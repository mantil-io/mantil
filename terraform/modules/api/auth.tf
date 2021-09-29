resource "aws_lambda_function" "authorizer" {
  count         = var.authorizer == null ? 0 : 1
  role          = aws_iam_role.authorizer[0].arn
  s3_bucket     = var.functions_bucket
  s3_key        = var.authorizer.s3_key
  function_name = "${var.name_prefix}-authorizer"
  handler       = "bootstrap"
  runtime       = "provided.al2"
  environment {
    variables = {
      MANTIL_PUBLIC_KEY = var.authorizer.public_key
    }
  }
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
  name             = "${var.name_prefix}-ws-authorizer"
}

resource "aws_apigatewayv2_authorizer" "http" {
  count                             = var.authorizer == null ? 0 : 1
  api_id                            = aws_apigatewayv2_api.http.id
  authorizer_type                   = "REQUEST"
  authorizer_uri                    = aws_lambda_function.authorizer[0].invoke_arn
  identity_sources                  = ["$request.header.${var.authorizer.authorization_header}"]
  authorizer_payload_format_version = "1.0"
  name                              = "${var.name_prefix}-http-authorizer"
}
