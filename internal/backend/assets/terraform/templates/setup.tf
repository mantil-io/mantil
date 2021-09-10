locals {
  aws_region       = "eu-central-1"
  functions_bucket = "mantil-downloads" # bucket with backend functions
  project_bucket   = "{{.Bucket}}"      # TODO bucket for backend configuration/state (created in advance)
  functions = {
    "init" = {
      s3_key      = "functions/init.zip"
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
      s3_key      = "functions/data.zip"
      memory_size = 128,
      timeout     = 900
    },
    "security" = {
      s3_key      = "functions/security.zip"
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

  skip_get_ec2_platforms = true
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

