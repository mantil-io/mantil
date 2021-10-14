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
  global_env = {
    env1 = "value1"
    env2 = "value2"
  }
}

terraform {
  backend "s3" {
    bucket = "bucket-name"
    key    = "bucket-prefix/terraform/state.tfstate"
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
  prefix     = "mantil-project-${local.project_name}"
  suffix      = "abcdef"
  global_env = local.global_env
}

module "public_site" {
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
  project_name = local.project_name
  integrations = concat(
  [ for f in module.functions.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
      integration_method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
    }
  ],
  [
    {
      type : "HTTP_PROXY"
      method : "GET"
      integration_method: "GET"
      route : "/public"
      uri : "http://${module.public_site.url}"
    }
  ])
}

output "url" {
  value = module.api.http_url
}

output "functions_bucket" {
  value = local.project_bucket
}

output "public_site_bucket" {
  value = module.public_site.bucket
}

output "ws_url" {
  value = module.api.ws_url
}
