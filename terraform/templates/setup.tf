locals {
  aws_region       = "{{.Region}}"
  functions_bucket = "{{.FunctionsBucket}}" # bucket with backend functions
  project_bucket   = "{{.Bucket}}"          # bucket for backend configuration/state
  functions = {
    "init" = {
      s3_key      = "{{.FunctionsPath}}/init.zip"
      memory_size = 128
      timeout     = 900
    },
    "deploy" = {
      s3_key      = "{{.FunctionsPath}}/deploy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:{{.Region}}:553035198032:layer:git-lambda2:8", "arn:aws:lambda:{{.Region}}:477361877445:layer:terraform-lambda:1"]
    },
    "data" = {
      s3_key      = "{{.FunctionsPath}}/data.zip"
      memory_size = 128,
      timeout     = 900
    },
    "security" = {
      s3_key      = "{{.FunctionsPath}}/security.zip"
      memory_size = 128,
      timeout     = 900,
    },
    "destroy" = {
      s3_key      = "{{.FunctionsPath}}/destroy.zip"
      memory_size = 512,
      timeout     = 900
      layers      = ["arn:aws:lambda:{{.Region}}:553035198032:layer:git-lambda2:8", "arn:aws:lambda:{{.Region}}:477361877445:layer:terraform-lambda:1"]
    }
  }
  ws_handler = {
    name        = "ws-handler"
    s3_key      = "{{.FunctionsPath}}/ws-handler.zip"
    memory_size = 128
    timeout     = 900
  }
  ws_sqs_forwarder = {
    name        = "ws-sqs-forwarder"
    s3_key      = "{{.FunctionsPath}}/ws-sqs-forwarder.zip"
    memory_size = 128
    timeout     = 900
  }
}

terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "{{.BucketPrefix}}terraform/state.tfstate"
    region = "{{.Region}}"
  }
}

provider "aws" {
  region                 = "{{.Region}}"
  skip_get_ec2_platforms = true
}

module "funcs" {
  source    = "../modules/backend-funcs"
  functions = local.functions
  s3_bucket = local.functions_bucket
}

module "iam" {
  source           = "../modules/backend-iam"
  backend_role_arn = module.funcs.role_arn
}

module "api" {
  source           = "../modules/api"
  name_prefix      = "mantil"
  functions_bucket = local.functions_bucket
  integrations = [for f in module.funcs.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
      enable_auth : false,
    }
  ]
  authorizer = {
    authorization_header = "X-Mantil-Access-Token"
    public_key           = "{{ .PublicKey }}"
    s3_key               = "{{.FunctionsPath}}/authorizer.zip"
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
