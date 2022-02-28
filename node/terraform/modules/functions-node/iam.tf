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
      "lambda:DeleteFunction",
      "lambda:RemovePermission",
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
      "logs:DeleteLogGroup",
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
      "iam:DeleteRole",
      "iam:DeleteRolePolicy",
      "iam:ListInstanceProfilesForRole",
    ]
    resources = [
      "arn:aws:iam::*:role/*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "iam:CreateServiceLinkedRole",
    ]
    resources = [
      "arn:aws:iam::*:role/aws-service-role/ops.apigateway.amazonaws.com/AWSServiceRoleForAPIGateway",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:CreateBucket",
      "s3:GetBucketObjectLockConfiguration",
      "s3:GetBucketAcl",
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
      "arn:aws:s3:::*-${var.suffix}/*",
      "arn:aws:s3:::*-${var.suffix}",
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
  statement {
    effect = "Allow"
    actions = [
      "events:TagResource",
      "events:ListTagsForResource",
      "events:PutRule",
      "events:DescribeRule",
      "events:DeleteRule",
      "events:PutTargets",
      "events:ListTargetsByRule",
      "events:RemoveTargets",
    ]
    resources = [
      "arn:aws:events:*:*:rule/*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "acm:ListCertificates",
      "acm:DescribeCertificate",
      "acm:ListTagsForCertificate",
    ]
    resources = [
      "*",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "route53:ListHostedZones",
      "route53:GetHostedZone",
      "route53:ListTagsForResource",
      "route53:ChangeResourceRecordSets",
      "route53:GetChange",
      "route53:ListResourceRecordSets",
    ]
    resources = [
      "*",
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
    resources = [
      "arn:aws:logs:*:*:log-group:*-${var.suffix}",
      "arn:aws:logs:*:*:log-group:*-${var.suffix}:log-stream:*",
    ]
  }
}

data "aws_iam_policy_document" "destroy" {
  statement {
    effect = "Allow"
    actions = [
      "apigateway:GET",
      "apigateway:POST",
      "apigateway:PATCH",
      "apigateway:DELETE",
      "apigateway:PUT",
    ]
    resources = [
      "*",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "lambda:DeleteFunction",
      "lambda:GetFunction",
      "lambda:GetFunctionCodeSigningConfig",
      "lambda:GetPolicy",
      "lambda:ListVersionsByFunction",
      "lambda:RemovePermission",
    ]
    resources = [
      "arn:aws:lambda:*:*:function:*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DeleteLogGroup",
      "logs:ListTagsLogGroup",
    ]
    resources = [
      "arn:aws:logs:*:*:log-group:*-${var.suffix}",
      "arn:aws:logs:*:*:log-group:*-${var.suffix}:log-stream:*",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups",
      "logs:ListLogDeliveries",
      "logs:DeleteLogDelivery",
    ]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "iam:DeleteRole",
      "iam:DeleteRolePolicy",
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:ListAttachedRolePolicies",
      "iam:ListInstanceProfilesForRole",
      "iam:ListRolePolicies",
    ]
    resources = [
      "arn:aws:iam::*:role/*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:DeleteObjectVersion",
      "s3:DeleteBucketPolicy",
      "s3:GetAccelerateConfiguration",
      "s3:GetBucketCORS",
      "s3:GetBucketLocation",
      "s3:GetBucketLogging",
      "s3:GetBucketObjectLockConfiguration",
      "s3:GetBucketAcl",
      "s3:GetBucketPolicy",
      "s3:GetBucketRequestPayment",
      "s3:GetBucketTagging",
      "s3:GetBucketVersioning",
      "s3:GetBucketWebsite",
      "s3:GetEncryptionConfiguration",
      "s3:GetLifecycleConfiguration",
      "s3:GetReplicationConfiguration",
      "s3:DeleteBucketWebsite",
    ]
    resources = [
      "arn:aws:s3:::*-${var.suffix}/*",
      "arn:aws:s3:::*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DeleteTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ListTagsOfResource",
      "dynamodb:DescribeTimeToLive",
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "tag:GetResources"
    ]
    resources = [
      "*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "events:ListTagsForResource",
      "events:DescribeRule",
      "events:DeleteRule",
      "events:ListTargetsByRule",
      "events:RemoveTargets",
    ]
    resources = [
      "arn:aws:events:*:*:rule/*-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "route53:GetHostedZone",
      "route53:ChangeResourceRecordSets",
      "route53:GetChange",
      "route53:ListResourceRecordSets",
    ]
    resources = [
      "*",
    ]
  }
}

data "aws_iam_policy_document" "auth" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ListTagsOfResource",
      "dynamodb:TagResource",
      "dynamodb:DescribeTimeToLive",
      "dynamodb:CreateTable",
      "dynamodb:Query",
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:BatchWriteItem",
      "dynamodb:BatchGetItem",
      "dynamodb:DeleteItem",
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/mantil-kv-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
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
      "ssm:GetParameter",
    ]
    resources = [
      "*",
    ]
  }
}

data "aws_iam_policy_document" "node" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ListTagsOfResource",
      "dynamodb:TagResource",
      "dynamodb:DescribeTimeToLive",
      "dynamodb:CreateTable",
      "dynamodb:Query",
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:BatchWriteItem",
      "dynamodb:BatchGetItem",
      "dynamodb:DeleteItem",
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/mantil-kv-${var.suffix}",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
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
      "ssm:GetParameter",
    ]
    resources = [
      "*",
    ]
  }
}
