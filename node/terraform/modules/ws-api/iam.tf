resource "aws_iam_role" "ws_handler" {
  name = format(var.naming_template, "ws-handler")
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "ws_handler" {
  statement {
    effect    = "Allow"
    actions   = ["lambda:InvokeFunction"]
    resources = ["arn:aws:lambda:*:*:function:${format(var.naming_template, "*")}"]
  }
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:Query",
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:DescribeTable",
      "dynamodb:BatchWriteItem",
      "dynamodb:BatchGetItem",
      "dynamodb:DeleteItem",
    ]
    resources = ["arn:aws:dynamodb:*:*:table/${local.dynamodb_table}"]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:log-group:*-${var.suffix}",
      "arn:aws:logs:*:*:log-group:*-${var.suffix}:log-stream:*",
    ]
  }
}

resource "aws_iam_role_policy" "ws_handler" {
  name   = format(var.naming_template, "ws-handler")
  role   = aws_iam_role.ws_handler.id
  policy = data.aws_iam_policy_document.ws_handler.json
}

resource "aws_iam_role" "ws_forwarder" {
  name = format(var.naming_template, "ws-forwarder")
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "ws_forwarder" {
  statement {
    effect    = "Allow"
    actions   = ["execute-api:ManageConnections"]
    resources = ["arn:aws:execute-api:*:*:${aws_apigatewayv2_api.ws.id}/*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:Query",
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:DescribeTable",
      "dynamodb:BatchWriteItem",
      "dynamodb:BatchGetItem",
      "dynamodb:DeleteItem",
    ]
    resources = ["arn:aws:dynamodb:*:*:table/${local.dynamodb_table}"]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:log-group:*-${var.suffix}",
      "arn:aws:logs:*:*:log-group:*-${var.suffix}:log-stream:*",
    ]
  }
}

resource "aws_iam_role_policy" "ws_forwarder" {
  name   = format(var.naming_template, "ws-forwarder")
  role   = aws_iam_role.ws_forwarder.id
  policy = data.aws_iam_policy_document.ws_forwarder.json
}
