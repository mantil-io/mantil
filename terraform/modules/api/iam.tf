resource "aws_iam_role" "authorizer" {
  count = var.authorizer == null ? 0 : 1
  name  = "${var.name_prefix}-authorizer"
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

// TODO permissions
resource "aws_iam_role_policy" "authorizer" {
  count = var.authorizer == null ? 0 : 1
  name  = "${var.name_prefix}-authorizer"
  role  = aws_iam_role.authorizer[0].id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role" "ws_handler" {
  name = "${var.name_prefix}-ws-handler"
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

// TODO permissions
resource "aws_iam_role_policy" "ws_handler" {
  name = "${var.name_prefix}-ws-handler"
  role = aws_iam_role.ws_handler.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role" "sqs_forwarder" {
  name = "${var.name_prefix}-ws-sqs-forwarder"
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

// TODO permissions
resource "aws_iam_role_policy" "sqs_forwarder" {
  name = "${var.name_prefix}-ws-sqs-forwarder"
  role = aws_iam_role.sqs_forwarder.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
      }
    ]
  })
}
