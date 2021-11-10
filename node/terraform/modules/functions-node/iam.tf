//data "aws_iam_policy_document" "deploy" {}

data "aws_iam_policy_document" "security" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole"
    ]
    resources = [var.cli_role_arn]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["*"]
  }
}

//data "aws_iam_policy_document" "destroy" {}
