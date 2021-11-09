locals {
  name = "${var.prefix}-cli-user-${var.suffix}"
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
    resources = ["arn:aws:s3:::${var.prefix}-*-${var.suffix}/*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogStreams",
      "logs:FilterLogEvents"
    ]
    resources = ["*"]
    dynamic "condition" {
      for_each = var.tags
      content {
        test     = "StringEquals"
        variable = "aws:resourceTag/${condition.key}"
        values   = ["${condition.value}"]
      }
    }
  }
}

resource "aws_iam_role_policy" "cli_role" {
  name   = local.name
  role   = aws_iam_role.cli_role.id
  policy = data.aws_iam_policy_document.cli_role.json
}
