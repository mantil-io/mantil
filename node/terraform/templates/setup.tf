locals {
  aws_region        = "{{.Region}}"
  functions_bucket  = "{{.FunctionsBucket}}" # bucket with backend functions
  functions_s3_path = "{{.FunctionsPath}}"
  project_bucket    = "{{.Bucket}}" # bucket for backend configuration/state
  tags = {
    {{- range $key, $value := .ResourceTags}}
    {{$key}} = "{{$value}}"
    {{- end}}
  }
}

terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "{{.BucketPrefix}}/state.tfstate"
    region = "{{.Region}}"
  }
}

provider "aws" {
  region                 = "{{.Region}}"
  skip_get_ec2_platforms = true

  default_tags {
    tags = local.tags
  }
}

module "functions" {
  source           = "../../modules/functions-node"
  functions_bucket = local.functions_bucket
  functions_path   = local.functions_s3_path
  suffix           = "{{.ResourceSuffix}}"
  region           = local.aws_region
  cli_role_arn     = module.cli_role.arn
}


module "cli_role" {
  source = "../../modules/cli-role"
  prefix = "mantil"
  suffix = "{{.ResourceSuffix}}"
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
