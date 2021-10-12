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
  static_websites = {
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
}

module "funcs" {
  source          = "../../modules/funcs"
  project_name    = local.project_name
  functions       = local.functions
  s3_bucket       = local.project_bucket
  static_websites = local.static_websites
  global_env      = local.global_env
}

module "api" {
  source = "../../modules/api"
  prefix = "${local.project_name}"
  suffix = "123456" // TODO stage uuid
  functions_bucket = local.functions_bucket
  functions_s3_path = local.functions_s3_path
  project_name = local.project_name
  integrations = concat(
  [ for f in module.funcs.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
      integration_method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
    }
  ],
  [ for w in module.funcs.static_websites :
    {
      type : "HTTP_PROXY"
      method : "GET"
      integration_method: "GET"
      route : "/public/${w.name}"
      uri : "http://${w.url}"
    }
  ])
}

output "url" {
  value = module.api.http_url
}

output "functions_bucket" {
  value = local.project_bucket
}

output "static_websites" {
  value = module.funcs.static_websites
}

output "ws_url" {
  value = module.api.ws_url
}
