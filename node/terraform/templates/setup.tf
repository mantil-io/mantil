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
  ssm_prefix = "/mantil-node-{{.ResourceSuffix}}"
}

terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "{{.BucketPrefix}}/state.tfstate"
    region = "{{.Region}}"
  }
}

provider "aws" {
  version                = "~> 4.0"
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
  naming_template  = "{{.NamingTemplate}}"
  ssm_prefix       = local.ssm_prefix
}


module "cli_role" {
  source = "../../modules/cli-role"
  suffix = "{{.ResourceSuffix}}"
  naming_template = "{{.NamingTemplate}}"
}

module "api" {
  source            = "../../modules/api"
  suffix            = "{{.ResourceSuffix}}"
  naming_template   = "{{.NamingTemplate}}"
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
    env = {
      {{- range $key, $value := .AuthEnv}}
      {{$key}} = "{{$value}}"
      {{- end}}
    }
  }
}

resource "aws_ssm_parameter" "public_key" {
  name  = "${local.ssm_prefix}/public_key"
  type  = "String"
  value = "{{.PublicKey}}"
  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "private_key" {
  name  = "${local.ssm_prefix}/private_key"
  type  = "SecureString"
  value = "{{.PrivateKey}}"
  lifecycle {
    ignore_changes = [value]
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
