resource "aws_apigatewayv2_api" "http" {
  name          = "${var.prefix}-http-${var.suffix}"
  protocol_type = "HTTP"
  cors_configuration {
    allow_origins = toset(["*"])
  }
}

resource "aws_cloudwatch_log_group" "http_access_logs" {
  name              = "${var.prefix}-http-access-logs-${var.suffix}"
  retention_in_days = 14
}

resource "aws_apigatewayv2_stage" "http_default" {
  name          = "$default"
  api_id        = aws_apigatewayv2_api.http.id
  deployment_id = aws_apigatewayv2_deployment.http.id

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.http_access_logs.arn
    format          = "{ \"requestId\":\"$context.requestId\", \"ip\": \"$context.identity.sourceIp\", \"requestTime\":\"$context.requestTime\", \"httpMethod\":\"$context.httpMethod\",\"routeKey\":\"$context.routeKey\", \"status\":\"$context.status\",\"protocol\":\"$context.protocol\", \"responseLength\":\"$context.responseLength\" }"
  }

  default_route_settings {
    detailed_metrics_enabled = true
    throttling_burst_limit   = 100
    throttling_rate_limit    = 500
  }
}

resource "aws_apigatewayv2_route" "http" {
  for_each           = local.integrations
  api_id             = aws_apigatewayv2_api.http.id
  route_key          = "${each.value.method} ${each.value.route}"
  target             = "integrations/${aws_apigatewayv2_integration.http[each.key].id}"
  authorization_type = each.value.enable_auth ? "CUSTOM" : "NONE"
  authorizer_id      = each.value.enable_auth ? aws_apigatewayv2_authorizer.http[0].id : null
}

resource "aws_apigatewayv2_integration" "http" {
  for_each           = local.integrations
  api_id             = aws_apigatewayv2_api.http.id
  integration_type   = each.value.type
  integration_method = each.value.integration_method
  integration_uri    = each.value.uri
}

resource "aws_apigatewayv2_route" "http_proxy" {
  for_each           = local.integrations
  api_id             = aws_apigatewayv2_api.http.id
  route_key          = "${each.value.method} ${each.value.route}/{proxy+}"
  target             = "integrations/${aws_apigatewayv2_integration.http_proxy[each.key].id}"
  authorization_type = each.value.enable_auth ? "CUSTOM" : "NONE"
  authorizer_id      = each.value.enable_auth ? aws_apigatewayv2_authorizer.http[0].id : null
}

resource "aws_apigatewayv2_integration" "http_proxy" {
  for_each           = local.integrations
  api_id             = aws_apigatewayv2_api.http.id
  integration_type   = each.value.type
  integration_method = each.value.integration_method
  integration_uri    = each.value.type == "AWS_PROXY" ? each.value.uri : "${each.value.uri}/{proxy}"
  request_parameters = {
    "overwrite:path" = "$request.path.proxy"
  }
}

resource "aws_apigatewayv2_deployment" "http" {
  depends_on = [
    aws_apigatewayv2_route.http,
    aws_apigatewayv2_integration.http,
    aws_apigatewayv2_route.http_proxy,
    aws_apigatewayv2_integration.http_proxy
  ]
  api_id = aws_apigatewayv2_api.http.id
  triggers = {
    redeployment = sha1(jsonencode([
      aws_apigatewayv2_route.http,
      aws_apigatewayv2_integration.http,
      aws_apigatewayv2_route.http_proxy,
      aws_apigatewayv2_integration.http_proxy,
      local.integrations
    ]))
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lambda_permission" "api_gateway_invoke" {
  for_each      = { for k, v in local.integrations : k => v if v.type == "AWS_PROXY" }
  function_name = each.value.lambda_name
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http.execution_arn}/*/*"
}
