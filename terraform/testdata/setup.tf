locals {
  aws_region        = "aws-region"
  functions_bucket  = "functions-bucket" # bucket with backend functions
  functions_s3_path = "functions-path"
  project_bucket    = "bucket-name" # bucket for backend configuration/state
  functions = {
    "deploy" = {
      s3_key      = "functions-path/deploy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:aws-region:553035198032:layer:git-lambda2:8", "arn:aws:lambda:aws-region:477361877445:layer:terraform-lambda:1"]
    },
    "data" = {
      s3_key      = "functions-path/data.zip"
      memory_size = 128,
      timeout     = 900
    },
    "security" = {
      s3_key      = "functions-path/security.zip"
      memory_size = 128,
      timeout     = 900,
    },
    "destroy" = {
      s3_key      = "functions-path/destroy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:aws-region:553035198032:layer:git-lambda2:8", "arn:aws:lambda:aws-region:477361877445:layer:terraform-lambda:1"]
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
}

module "funcs" {
  source    = "../../modules/backend-funcs"
  functions = local.functions
  s3_bucket = local.functions_bucket
}

module "iam" {
  source           = "../../modules/backend-iam"
  backend_role_arn = module.funcs.role_arn
}

module "api" {
  source            = "../../modules/api"
  name_prefix       = "mantil"
  functions_bucket  = local.functions_bucket
  functions_s3_path = local.functions_s3_path
  integrations = [for f in module.funcs.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
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
  value = module.iam.cli_role
}

output "ws_url" {
  value = module.api.ws_url
}
