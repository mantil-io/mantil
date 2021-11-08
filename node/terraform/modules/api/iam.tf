resource "aws_iam_role" "authorizer" {
  count = var.authorizer == null ? 0 : 1
  name  = "${var.prefix}-authorizer-${var.suffix}"
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
  name  = "${var.prefix}-authorizer-${var.suffix}"
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
