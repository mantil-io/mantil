locals {
  aws_region        = "aws-region"
  functions_bucket  = "functions-bucket" # bucket with backend functions
  functions_s3_path = "functions-path"
  project_bucket    = "bucket-name" # bucket for backend configuration/state
  functions = {
    "deploy" = {
      method      = "POST"
      s3_key      = "functions-path/deploy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:aws-region:477361877445:layer:terraform-lambda:1"]
    },
    "security" = {
      method      = "GET"
      s3_key      = "functions-path/security.zip"
      memory_size = 128,
      timeout     = 900,
    },
    "destroy" = {
      method      = "POST"
      s3_key      = "functions-path/destroy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:aws-region:477361877445:layer:terraform-lambda:1"]
    }
  }
}

terraform {
  backend "s3" {
    bucket = "bucket-name"
    key    = "setup/terraform/state.tfstate"
    region = "aws-region"
  }
}

provider "aws" {
  region                 = "aws-region"
  skip_get_ec2_platforms = true

  default_tags {
    tags = {
      tag1 = "value1"
      tag2 = "value2"
    }
  }
}

module "functions" {
  source    = "../../modules/functions"
  functions = local.functions
  s3_bucket = local.functions_bucket
  prefix    = "mantil"
  suffix    = "abcdef"
}

module "cli_role" {
  source           = "../../modules/iam-role"
  name             = "mantil-cli-user-abcdef"
}

module "api" {
  source            = "../../modules/api"
  prefix            = "mantil"
  suffix            = "abcdef"
  functions_bucket  = local.functions_bucket
  functions_s3_path = local.functions_s3_path
  integrations = [for f in module.functions.functions :
    {
      type : "AWS_PROXY"
      method : local.functions[f.name].method
      integration_method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
      enable_auth : true,
    }
  ]
  authorizer = {
    authorization_header = "Authorization"
    public_key           = "public-key"
  }
}

# expose aws region and profile for use in shell scripts
output "aws_region" {
  value = local.aws_region
}

output "project_bucket" {
  value = local.project_bucket
}

output "url" {
  value = module.api.http_url
}

output "cli_role" {
  value = module.cli_role.arn
}

output "ws_url" {
  value = module.api.ws_url
}
