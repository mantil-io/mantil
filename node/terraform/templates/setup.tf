locals {
  aws_region        = "{{.Region}}"
  functions_bucket  = "{{.FunctionsBucket}}" # bucket with backend functions
  functions_s3_path = "{{.FunctionsPath}}"
  project_bucket    = "{{.Bucket}}" # bucket for backend configuration/state
  functions = {
    "deploy" = {
      method      = "POST"
      s3_key      = "{{.FunctionsPath}}/deploy.zip"
      memory_size = 512,
      timeout     = 900
      architecture = "arm64"
      layers      = ["arn:aws:lambda:{{.Region}}:477361877445:layer:terraform-lambda:3"]
    },
    "security" = {
      method      = "GET"
      s3_key      = "{{.FunctionsPath}}/security.zip"
      memory_size = 128,
      timeout     = 900,
      architecture = "arm64"
    },
    "destroy" = {
      method      = "POST"
      s3_key      = "{{.FunctionsPath}}/destroy.zip"
      memory_size = 512,
      timeout     = 900
      architecture = "arm64"
      layers      = ["arn:aws:lambda:{{.Region}}:477361877445:layer:terraform-lambda:3"]
    }
  }
}

terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "{{.BucketPrefix}}/terraform/state.tfstate"
    region = "{{.Region}}"
  }
}

provider "aws" {
  region                 = "{{.Region}}"
  skip_get_ec2_platforms = true

  default_tags {
    tags = {
      {{- range $key, $value := .ResourceTags}}
      {{$key}} = "{{$value}}"
      {{- end}}
    }
  }
}

module "functions" {
  source    = "../../modules/functions"
  functions = local.functions
  s3_bucket = local.functions_bucket
  prefix    = "mantil"
  suffix    = "{{.ResourceSuffix}}"
}

module "cli_role" {
  source           = "../../modules/iam-role"
  name             = "mantil-cli-user-{{.ResourceSuffix}}"
}

module "api" {
  source            = "../../modules/api"
  prefix            = "mantil"
  suffix            = "{{.ResourceSuffix}}"
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
    env = {
      {{- range $key, $value := .AuthEnv}}
      {{$key}} = "{{$value}}"
      {{- end}}
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

output "ws_url" {
  value = module.api.ws_url
}
