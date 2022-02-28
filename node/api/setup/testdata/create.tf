locals {
  aws_region        = "aws-region"
  functions_bucket  = "functions-bucket" # bucket with backend functions
  functions_s3_path = "functions-path"
  project_bucket    = "bucket-name" # bucket for backend configuration/state
  tags = {
    tag1 = "value1"
    tag2 = "value2"
  }
  ssm_prefix = "/mantil-node-abcdef"
  auth_env = {
    publicKey = "public_key"
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
  version                = "~> 4.0"
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
  naming_template  = "mantil-%s"
  auth_env         = local.auth_env
}


module "cli_role" {
  source = "../../modules/cli-role"
  suffix = "abcdef"
  naming_template = "mantil-%s"
}

module "api" {
  source            = "../../modules/api"
  suffix            = "abcdef"
  naming_template   = "mantil-%s"
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
      enable_auth : f.name != "auth" ? true : false,
    }
  ]
  authorizer = {
    authorization_header = "Authorization"
    env = local.auth_env
  }
}

module "params" {
  source = "../../modules/params"
  path_prefix = local.ssm_prefix
  params = [
    {
      name : "public_key"
      value : "public_key"
    },
    {
      name : "private_key"
      value : "private_key"
      secure : true
    },
    {
      name: "github_org"
      value : "github_org"
    },
    {
      name: "github_user"
      value : ""
    }
  ]
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
