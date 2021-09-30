locals {
  ws_lambda_env = {
    "MANTIL_KV_TABLE_NAME" : local.dynamodb_table,
    "MANTIL_PROJECT_NAME" : var.project_name
  }
  ws_handler = {
    name   = "${var.name_prefix}-ws-handler"
    s3_key = "functions/ws-handler.zip"
  }
  sqs_forwarder = {
    name   = "${var.name_prefix}-sqs-forwarder"
    s3_key = "functions/ws-sqs-forwarder.zip"
  }
  dynamodb_table = "${var.name_prefix}-ws-connections"
}

resource "aws_apigatewayv2_api" "ws" {
  name          = "${var.name_prefix}-ws"
  protocol_type = "WEBSOCKET"

  route_selection_expression = "\\$default"
}

resource "aws_apigatewayv2_stage" "ws_default" {
  name          = "$default"
  api_id        = aws_apigatewayv2_api.ws.id
  deployment_id = aws_apigatewayv2_deployment.ws.id
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
  depends_on = [aws_apigatewayv2_route.ws_handler]
  api_id     = aws_apigatewayv2_api.ws.id
  triggers = {
    redeployment = sha1(jsonencode([
      aws_apigatewayv2_api.ws.body
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

  environment {
    variables = local.ws_lambda_env
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

resource "aws_lambda_function" "sqs_forwarder" {
  role = aws_iam_role.sqs_forwarder.arn

  s3_bucket = var.functions_bucket
  s3_key    = local.sqs_forwarder.s3_key

  function_name = local.sqs_forwarder.name
  handler       = "runtime"
  runtime       = "provided.al2"

  environment {
    variables = local.ws_lambda_env
  }
}

resource "aws_cloudwatch_log_group" "sqs_forwarder_log_group" {
  name              = "/aws/lambda/${local.sqs_forwarder.name}"
  retention_in_days = 14
}

resource "aws_sqs_queue" "queue" {
  name                        = "${var.name_prefix}-ws-queue.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
  visibility_timeout_seconds  = aws_lambda_function.sqs_forwarder.timeout
}

resource "aws_lambda_event_source_mapping" "handler_trigger" {
  event_source_arn = aws_sqs_queue.queue.arn
  function_name    = aws_lambda_function.sqs_forwarder.arn
  batch_size       = 10
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
