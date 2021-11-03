locals {
  aws_region       = "{{.Region}}"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  project_name     = "{{.Project}}-{{.Stage}}"
  project_bucket   = "{{.Bucket}}"                           # TODO bucket for project configuration/state/functions (created in advance)
  functions_bucket = "{{.AccountFunctionsBucket}}"
  functions_s3_path = "{{.AccountFunctionsPath}}"
  functions = {
    {{- range .Functions}}
    {{.Name}} = {
      s3_key = "{{.S3Key}}"
      runtime = "{{.Runtime}}"
      memory_size = {{.MemorySize}}
      handler = "{{.Handler}}"
      timeout = {{.Timeout}}
      env = {
        {{- range $key, $value := .Env}}
        {{$key}} = "{{$value}}"
        {{- end}}
      }
    }
    {{- end}}
  }
  ws_env = {
    {{- range $key, $value := .WsEnv}}
    {{$key}} = "{{$value}}"
    {{- end}}
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
  region = local.aws_region
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
  source     = "../../modules/functions"
  functions  = local.functions
  s3_bucket  = local.project_bucket
  prefix     = "${local.project_name}"
  suffix      = "{{.ResourceSuffix}}"
}

module "public_site" {
  source = "../../modules/public-site"
  prefix = "mantil-public-${local.project_name}"
  suffix = "{{.ResourceSuffix}}"
}

module "api" {
  source = "../../modules/api"
  prefix = "${local.project_name}"
  suffix = "{{.ResourceSuffix}}"
  functions_bucket = local.functions_bucket
  functions_s3_path = local.functions_s3_path
  ws_enabled = true
  integrations = concat(
  [ for f in module.functions.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
      integration_method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
    }
  ],
  [
    {
      type : "HTTP_PROXY"
      method : "GET"
      integration_method: "GET"
      route : "/public"
      uri : "http://${module.public_site.url}"
    }
  ])
  ws_env = local.ws_env
}

output "url" {
  value = module.api.http_url
}

output "functions_bucket" {
  value = local.project_bucket
}

output "public_site_bucket" {
  value = module.public_site.bucket
}

output "ws_url" {
  value = module.api.ws_url
}
