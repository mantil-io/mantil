locals {
  name = format(var.naming_template, "cli-user")
}

resource "aws_iam_role" "cli_role" {
  name = local.name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole",
        Principal = {
          AWS = "*",
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "cli_role" {
  statement {
    effect    = "Allow"
    actions   = ["s3:PutObject"]
    resources = ["arn:aws:s3:::*-${var.suffix}/*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogStreams",
      "logs:FilterLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:log-group:*-${var.suffix}",
      "arn:aws:logs:*:*:log-group:*-${var.suffix}:log-stream:*",
    ]
  }
}

resource "aws_iam_role_policy" "cli_role" {
  name   = local.name
  role   = aws_iam_role.cli_role.id
  policy = data.aws_iam_policy_document.cli_role.json
}
