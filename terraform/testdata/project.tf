locals {
  aws_region       = "aws-region"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  project_name     = "my-project-my-stage"
  project_bucket   = "bucket-name"                           # TODO bucket for project configuration/state/functions (created in advance)
  functions_bucket = "functions-bucket"
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
  ws_handler = {
    name        = "ws-handler"
    s3_key      = "functions-path/ws-handler.zip"
    memory_size = 128
    timeout     = 900
  }
  ws_sqs_forwarder = {
    name        = "ws-sqs-forwarder"
    s3_key      = "functions-path/ws-sqs-forwarder.zip"
    memory_size = 128
    timeout     = 900
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
}

module "api" {
  source = "../../modules/api"
  name_prefix = "mantil-project-${local.project_name}"
  functions_bucket = local.functions_bucket
  project_name = local.project_name
  integrations = concat(
  [ for f in module.funcs.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
    }
  ],
  [ for w in module.funcs.static_websites :
    {
      type : "HTTP_PROXY"
      method : "GET"
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
