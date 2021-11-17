locals {
  aws_region       = "aws-region"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  project_name     = "my-project-my-stage"
  project_bucket   = "bucket-name"                           # TODO bucket for project configuration/state/functions (created in advance)
  functions_bucket = "functions-bucket"
  functions_s3_path = "functions-path"
  functions = {
    function1 = {
      s3_key = "function1.zip"
      runtime = ""
      memory_size = 0
      handler = ""
      timeout = 0
      env = {
      }
    }
    function2 = {
      s3_key = "function2.zip"
      runtime = ""
      memory_size = 0
      handler = ""
      timeout = 0
      env = {
      }
    }
  }
  ws_env = {
    key = "value"
  }
  has_public = true
}

terraform {
  backend "s3" {
    bucket = "bucket-name"
    key    = "bucket-prefix/state.tfstate"
    region = "aws-region"
  }
}

provider "aws" {
  region = local.aws_region
  skip_get_ec2_platforms = true

  default_tags {
    tags = {
      tag1 = "value1"
      tag2 = "value2"
    }
  }
}

module "functions" {
  source     = "../../modules/functions"
  functions  = local.functions
  s3_bucket  = local.project_bucket
  prefix     = "${local.project_name}"
  suffix      = "abcdef"
}

module "public_site" {
  count  = local.has_public ? 1 : 0
  source = "../../modules/public-site"
  prefix = "mantil-public-${local.project_name}"
  suffix = "abcdef"
}

module "api" {
  source = "../../modules/api"
  prefix = "${local.project_name}"
  suffix = "abcdef"
  functions_bucket = local.functions_bucket
  functions_s3_path = local.functions_s3_path
  ws_enabled = true
  integrations = concat(
  [ for f in module.functions.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
      integration_method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn
      lambda_name : f.arn
    }
  ],
  [
    {
      type : "HTTP_PROXY"
      method : "GET"
      integration_method: "GET"
      route : "/public"
      uri : "http://${module.public_site[0].url}"
      is_default : true
    }
  ])
  ws_env = local.ws_env
}

output "url" {
  value = module.api.http_url
}

output "functions_bucket" {
  value = local.project_bucket
}

output "public_site_bucket" {
  value = local.has_public ? module.public_site[0].bucket : ""
}

output "ws_url" {
  value = module.api.ws_url
}
