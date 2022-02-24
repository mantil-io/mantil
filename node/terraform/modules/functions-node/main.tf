locals {
  functions = {
    "deploy" = {
      method       = "POST"
      s3_key       = "${var.functions_path}/deploy.zip"
      memory_size  = 1769
      timeout      = 900
      architecture = "arm64"
      layers       = ["arn:aws:lambda:${var.region}:477361877445:layer:terraform-layer:1"]
      policy       = data.aws_iam_policy_document.deploy.json
    },
    "security" = {
      method       = "GET"
      s3_key       = "${var.functions_path}/security.zip"
      memory_size  = 1769
      timeout      = 900
      architecture = "arm64"
      policy       = data.aws_iam_policy_document.security.json
    },
    "destroy" = {
      method       = "POST"
      s3_key       = "${var.functions_path}/destroy.zip"
      memory_size  = 1769
      timeout      = 900
      architecture = "arm64"
      layers       = ["arn:aws:lambda:${var.region}:477361877445:layer:terraform-layer:1"]
      policy       = data.aws_iam_policy_document.destroy.json
    }
    "auth" = {
      method       = "POST"
      s3_key       = "${var.functions_path}/auth.zip"
      memory_size  = 128
      timeout      = 900
      architecture = "arm64"
      policy       = data.aws_iam_policy_document.auth.json
      env          = var.auth_env
    }
  }
}

module "functions" {
  source          = "../functions"
  functions       = local.functions
  s3_bucket       = var.functions_bucket
  naming_template = var.naming_template
}
