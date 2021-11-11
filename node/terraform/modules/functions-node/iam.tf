data "aws_iam_policy_document" "deploy" {
  statement {
    effect = "Allow"
    actions = [
      "apigateway:GET",
      "apigateway:POST",
      "apigateway:PATCH",
      "apigateway:DELETE",
      "apigateway:PUT",
      "apigateway:TagResource"
    ]
    resources = [
      "*",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogDelivery",
      "logs:PutResourcePolicy",
      "logs:DescribeLogGroups",
      "logs:UpdateLogDelivery",
      "logs:GetLogDelivery",
      "logs:DescribeResourcePolicies",
      "logs:ListLogDeliveries",
    ]
    resources = [
      "*",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
    ]
    resources = [
      "arn:aws:s3:::mantil-releases*/*",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "lambda:CreateFunction",
      "lambda:GetFunctionConfiguration",
      "lambda:GetFunctionCodeSigningConfig",
      "lambda:UpdateFunctionCode",
      "lambda:ListVersionsByFunction",
      "lambda:GetFunction",
      "lambda:UpdateFunctionConfiguration",
      "lambda:AddPermission",
      "lambda:GetPolicy",
    ]
    resources = [
      "arn:aws:lambda:*:*:function:*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:CreateLogGroup",
      "logs:ListTagsLogGroup",
      "logs:PutRetentionPolicy",
      "logs:PutLogEvents",
    ]
    resources = [
      "arn:aws:logs:*:*:log-group:*-${var.suffix}",
      "arn:aws:logs:*:*:log-group:*-${var.suffix}:log-stream:*",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "iam:CreateRole",
      "iam:PutRolePolicy",
      "iam:ListAttachedRolePolicies",
      "iam:ListRolePolicies",
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:TagRole",
      "iam:PassRole",
    ]
    resources = [
      "arn:aws:iam::*:role/*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:CreateBucket",
      "s3:GetBucketObjectLockConfiguration",
      "s3:PutBucketAcl",
      "s3:GetBucketWebsite",
      "s3:GetReplicationConfiguration",
      "s3:PutObject",
      "s3:GetObject",
      "s3:GetLifecycleConfiguration",
      "s3:GetBucketTagging",
      "s3:GetBucketLogging",
      "s3:ListBucket",
      "s3:GetAccelerateConfiguration",
      "s3:GetBucketPolicy",
      "s3:GetEncryptionConfiguration",
      "s3:PutBucketTagging",
      "s3:GetBucketRequestPayment",
      "s3:GetBucketVersioning",
      "s3:PutBucketWebsite",
      "s3:GetBucketCORS",
      "s3:PutBucketPolicy",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:aws:s3:::mantil-*-${var.suffix}/*",
      "arn:aws:s3:::mantil-*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ListTagsOfResource",
      "dynamodb:TagResource",
      "dynamodb:DescribeTimeToLive",
      "dynamodb:CreateTable",
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/*-${var.suffix}",
    ]
  }
}

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
