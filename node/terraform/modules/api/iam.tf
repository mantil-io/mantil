resource "aws_iam_role" "authorizer" {
  count = var.authorizer == null ? 0 : 1
  name  = format(var.naming_template, "authorizer")
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

data "aws_iam_policy_document" "authorizer" {
  count = var.authorizer == null ? 0 : 1
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
  statement {
    effect = "Allow"
    actions = [
      "ssm:GetParameter",
    ]
    resources = [
      "*",
    ]
  }
}

resource "aws_iam_role_policy" "authorizer" {
  count  = var.authorizer == null ? 0 : 1
  name   = format(var.naming_template, "authorizer")
  role   = aws_iam_role.authorizer[0].id
  policy = data.aws_iam_policy_document.authorizer[0].json
}
