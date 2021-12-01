locals {
  aws_region       = "{{.Region}}"               # region where resources will be created (except cloudfront distribution which is global)
  project_bucket   = "{{.Bucket}}"               # bucket for project configuration/state/functions (created in advance)
  functions_bucket = "{{.NodeFunctionsBucket}}"
  functions_s3_path = "{{.NodeFunctionsPath}}"
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
  has_public = {{ .HasPublic }}
}

terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "{{.BucketPrefix}}/state.tfstate"
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
  naming_template = "{{.NamingTemplate}}"
}

module "public_site" {
  count  = local.has_public ? 1 : 0
  source = "../../modules/public-site"
  bucket_name = "{{.PublicBucketName}}"
}

module "api" {
  source = "../../modules/api"
  suffix = "{{.ResourceSuffix}}"
  naming_template = "{{.NamingTemplate}}"
  functions_bucket = local.functions_bucket
  functions_s3_path = local.functions_s3_path
  ws_enabled = true
  integrations = concat(
  [ for f in module.functions.functions :
    {
      type : "AWS_PROXY"
      method : "ANY"
      integration_method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn
      lambda_name : f.arn
    }
  ]{{if .HasPublic}},
  [
    {
      type : "HTTP_PROXY"
      method : "GET"
      integration_method: "GET"
      route : "/public"
      uri : "http://${module.public_site[0].url}"
      is_default : true
    }
  ]{{end}})
  ws_env = local.ws_env
}

output "url" {
  value = module.api.http_url
}

output "functions_bucket" {
  value = local.project_bucket
}

output "public_site_bucket" {
  value = local.has_public ? module.public_site[0].bucket : ""
}

output "ws_url" {
  value = module.api.ws_url
}
