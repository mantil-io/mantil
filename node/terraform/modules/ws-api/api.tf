resource "aws_apigatewayv2_api" "ws" {
  name          = "${var.prefix}-ws-${var.suffix}"
  protocol_type = "WEBSOCKET"

  route_selection_expression = "\\$default"
}

resource "aws_cloudwatch_log_group" "ws_access_logs" {
  name              = "${var.prefix}-ws-access-logs-${var.suffix}"
  retention_in_days = 14
}

resource "aws_apigatewayv2_stage" "ws_default" {
  name          = "$default"
  api_id        = aws_apigatewayv2_api.ws.id
  deployment_id = aws_apigatewayv2_deployment.ws.id

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.ws_access_logs.arn
    format          = "{ \"requestId\":\"$context.requestId\", \"ip\": \"$context.identity.sourceIp\", \"caller\":\"$context.identity.caller\", \"user\":\"$context.identity.user\",\"requestTime\":\"$context.requestTime\", \"eventType\":\"$context.eventType\",\"routeKey\":\"$context.routeKey\", \"status\":\"$context.status\",\"connectionId\":\"$context.connectionId\" }"
  }

  default_route_settings {
    data_trace_enabled       = true
    detailed_metrics_enabled = true
    logging_level            = "INFO"
    throttling_burst_limit   = 5000
    throttling_rate_limit    = 10000
  }
}

resource "aws_apigatewayv2_route" "ws_handler_connect" {
  api_id             = aws_apigatewayv2_api.ws.id
  route_key          = "$connect"
  target             = "integrations/${aws_apigatewayv2_integration.ws_handler.id}"
  authorization_type = var.authorizer == null ? "NONE" : "CUSTOM"
  authorizer_id      = var.authorizer == null ? null : aws_apigatewayv2_authorizer.ws[0].id
}

resource "aws_apigatewayv2_route" "ws_handler" {
  for_each = toset(["$disconnect", "$default"])

  api_id    = aws_apigatewayv2_api.ws.id
  route_key = each.key
  target    = "integrations/${aws_apigatewayv2_integration.ws_handler.id}"
}

resource "aws_apigatewayv2_integration" "ws_handler" {
  api_id           = aws_apigatewayv2_api.ws.id
  integration_type = "AWS_PROXY"

  integration_method = "POST"
  integration_uri    = aws_lambda_function.ws_handler.invoke_arn
}

resource "aws_apigatewayv2_deployment" "ws" {
  depends_on = [
    aws_apigatewayv2_route.ws_handler,
    aws_apigatewayv2_integration.ws_handler,
    aws_apigatewayv2_route.ws_handler_connect
  ]
  api_id = aws_apigatewayv2_api.ws.id
  triggers = {
    redeployment = sha1(jsonencode([
      aws_apigatewayv2_route.ws_handler,
      aws_apigatewayv2_integration.ws_handler,
      aws_apigatewayv2_route.ws_handler_connect
    ]))
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lambda_function" "ws_handler" {
  role = aws_iam_role.ws_handler.arn

  s3_bucket = var.functions_bucket
  s3_key    = local.ws_handler.s3_key

  function_name = local.ws_handler.name
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  environment {
    variables = local.ws_env
  }
}

resource "aws_cloudwatch_log_group" "ws_handler_log_group" {
  name              = "/aws/lambda/${local.ws_handler.name}"
  retention_in_days = 14
}

resource "aws_lambda_permission" "ws_handler_api_gateway_invoke" {
  function_name = local.ws_handler.name
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.ws.execution_arn}/*/*"
}

resource "aws_lambda_function" "ws_forwarder" {
  role = aws_iam_role.ws_forwarder.arn

  s3_bucket = var.functions_bucket
  s3_key    = local.ws_forwarder.s3_key

  function_name = local.ws_forwarder.name
  handler       = "runtime"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  environment {
    variables = local.ws_env
  }
}

resource "aws_cloudwatch_log_group" "ws_forwarder_log_group" {
  name              = "/aws/lambda/${local.ws_forwarder.name}"
  retention_in_days = 14
}

resource "aws_dynamodb_table" "table" {
  name      = local.dynamodb_table
  hash_key  = "PK"
  range_key = "SK"

  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }
}

resource "aws_lambda_permission" "authorizer_ws_api_gateway_invoke" {
  count         = var.authorizer == null ? 0 : 1
  function_name = var.authorizer.arn
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.ws.execution_arn}/*/*"
}

resource "aws_apigatewayv2_authorizer" "ws" {
  count            = var.authorizer == null ? 0 : 1
  api_id           = aws_apigatewayv2_api.ws.id
  authorizer_type  = "REQUEST"
  authorizer_uri   = var.authorizer.invoke_arn
  identity_sources = ["route.request.header.${var.authorizer.authorization_header}"]
  name             = "${var.prefix}-ws-authorizer-${var.suffix}"
}
