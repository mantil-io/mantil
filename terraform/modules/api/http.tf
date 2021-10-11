resource "aws_apigatewayv2_api" "http" {
  name          = "${var.name_prefix}-http"
  protocol_type = "HTTP"
  cors_configuration {
    allow_origins = toset(["*"])
  }
}

resource "aws_apigatewayv2_stage" "http_default" {
  name          = "$default"
  api_id        = aws_apigatewayv2_api.http.id
  deployment_id = aws_apigatewayv2_deployment.http.id
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
    aws_apigatewayv2_integration.http
  ]
  api_id = aws_apigatewayv2_api.http.id
  triggers = {
    redeployment = sha1(jsonencode([
      aws_apigatewayv2_api.http.body,
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
