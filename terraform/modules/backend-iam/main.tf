resource "aws_iam_role" "cli_user" {
  name        = "mantil-cli-user"
  description = "Role that will be used by mantil backend to issue temporary credentials for CLI users."

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole",
        Principal = {
          AWS = "${var.backend_role_arn}"
        }
      }
    ]
  })
}

// TODO permissions
resource "aws_iam_role_policy" "cli_user" {
  name = "mantil-cli-user"
  role = aws_iam_role.cli_user.id

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
