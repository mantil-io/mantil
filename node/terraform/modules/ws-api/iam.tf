resource "aws_iam_role" "ws_handler" {
  name = "${var.prefix}-ws-handler-${var.suffix}"
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
  name = "${var.prefix}-ws-handler-${var.suffix}"
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

resource "aws_iam_role" "ws_forwarder" {
  name = "${var.prefix}-ws-forwarder-${var.suffix}"
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
resource "aws_iam_role_policy" "ws_forwarder" {
  name = "${var.prefix}-ws-forwarder-${var.suffix}"
  role = aws_iam_role.ws_forwarder.id
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
