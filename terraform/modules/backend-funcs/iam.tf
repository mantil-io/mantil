resource "aws_iam_role" "lambda" {
  name = "${var.prefix}-${var.suffix}"
  //tags = { Name = "${local.name}-lambda" }

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

resource "aws_iam_instance_profile" "lambda" {
  name = "${var.prefix}-${var.suffix}"
  role = aws_iam_role.lambda.name
}

resource "aws_iam_role_policy" "lambda" {
  name = "${var.prefix}-${var.suffix}"
  role = aws_iam_role.lambda.id

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
