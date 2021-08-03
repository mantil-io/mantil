locals {
  aws_region       = "eu-central-1"
  functions_bucket = "mantil-downloads" # bucket with backend functions
  project_bucket   = "{{.Bucket}}"      # TODO bucket for backend configuration/state (created in advance)
  functions = {
    "init" = {
      s3_key      = "functions/init-59ecc75a02254965375b67d586901b107269c7dec5f8b889a5737fceb62a97c0.zip"
      memory_size = 128
      timeout     = 900
    },
    "deploy" = {
      s3_key      = "functions/deploy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:eu-central-1:553035198032:layer:git-lambda2:8", "arn:aws:lambda:eu-central-1:477361877445:layer:terraform-lambda:1"]
    },
    "data" = {
      s3_key      = "functions/data-54808399abfa95fbbfb1056dd0a9ed5073da147cb91bc85463d5cd9437f2e215.zip"
      memory_size = 128,
      timeout     = 900
    },
    "security" = {
      s3_key      = "functions/security-45341d9b230a65538e147276304ab4c8883be3dd593cb62e43c8854305f47d52.zip"
      memory_size = 128,
      timeout     = 900,
    },
    "destroy" = {
      s3_key      = "functions/destroy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:eu-central-1:553035198032:layer:git-lambda2:8", "arn:aws:lambda:eu-central-1:477361877445:layer:terraform-lambda:1"]
    }
  }
}

terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "terraform/state.tfstate"
    region = "eu-central-1"
  }
}

provider "aws" {
  region = local.aws_region
}

module "funcs" {
  source    = "http://localhost:8080/terraform/modules/backend-funcs.zip"
  functions = local.functions
  s3_bucket = local.functions_bucket
}

module "iam" {
  source           = "http://localhost:8080/terraform/modules/backend-iam.zip"
  backend_role_arn = module.funcs.role_arn
}

# expose aws region and profile for use in shell scripts
output "aws_region" {
  value = local.aws_region
}

output "functions" {
  value = module.funcs.functions
}

output "project_bucket" {
  value = local.project_bucket
}

output "url" {
  value = module.funcs.url
}

output "cli_role" {
  value = module.iam.cli_role
}
