locals {
  aws_region        = "aws-region"
  functions_bucket  = "functions-bucket" # bucket with backend functions
  functions_s3_path = "functions-path"
  project_bucket    = "bucket-name" # bucket for backend configuration/state
  tags = {
    tag1 = "value1"
    tag2 = "value2"
  }
}

terraform {
  backend "s3" {
    bucket = "bucket-name"
    key    = "setup/state.tfstate"
    region = "aws-region"
  }
}

provider "aws" {
  region                 = "aws-region"
  skip_get_ec2_platforms = true

  default_tags {
    tags = local.tags
  }
}

module "functions" {
  source           = "../../modules/functions-node"
  functions_bucket = local.functions_bucket
  functions_path   = local.functions_s3_path
  suffix           = "abcdef"
  region           = local.aws_region
  cli_role_arn     = module.cli_role.arn
}


module "cli_role" {
  source = "../../modules/cli-role"
  prefix = "mantil"
  suffix = "abcdef"
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
      method : f.method
      integration_method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
      enable_auth : true,
    }
  ]
  authorizer = {
    authorization_header = "Authorization"
    env = {
      publicKey = "key"
    }
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
