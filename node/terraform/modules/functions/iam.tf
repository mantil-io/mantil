resource "aws_iam_role" "lambda" {
  for_each = local.functions
  name     = each.value.function_name

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

resource "aws_iam_role_policy" "lambda" {
  for_each = local.functions
  name     = each.value.function_name
  role     = aws_iam_role.lambda[each.key].id
  policy   = each.value.policy
}
