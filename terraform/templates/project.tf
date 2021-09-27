locals {
  aws_region       = "{{.Region}}"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  project_name     = "{{.Name}}-{{.Stage}}"
  project_bucket   = "{{.Bucket}}"                           # TODO bucket for project configuration/state/functions (created in advance)
  functions_bucket = "{{.RuntimeFunctionsBucket}}"
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
  static_websites = {
    {{- range .PublicSites}}
    {{.Name}} = {
      name = "{{.Name}}"
    }
    {{- end}}
  }
  ws_handler = {
    name        = "ws-handler"
    s3_key      = "{{.RuntimeFunctionsPath}}/ws-handler.zip"
    memory_size = 128
    timeout     = 900
  }
  ws_sqs_forwarder = {
    name        = "ws-sqs-forwarder"
    s3_key      = "{{.RuntimeFunctionsPath}}/ws-sqs-forwarder.zip"
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
  region = local.aws_region

  skip_get_ec2_platforms = true
}

module "funcs" {
  source          = "http://localhost:8080/terraform/modules/funcs.zip"
  project_name    = local.project_name
  functions       = local.functions
  s3_bucket       = local.project_bucket
  static_websites = local.static_websites
}

module "api" {
  source = "http://localhost:8080/terraform/modules/api.zip"
  name_prefix = "mantil-project-${local.project_name}"
  functions_bucket = local.functions_bucket
  project_name = local.project_name
  integrations = concat(
  [ for f in module.funcs.functions :
    {
      type : "AWS_PROXY"
      method : "POST"
      route : "/${f.name}"
      uri : f.invoke_arn,
      lambda_name : f.arn,
    }
  ],
  [ for w in module.funcs.static_websites :
    {
      type : "HTTP_PROXY"
      method : "GET"
      route : "/public/${w.name}"
      uri : "http://${w.url}"
    }
  ])
}

output "url" {
  value = module.api.http_url
}

output "functions_bucket" {
  value = local.project_bucket
}

output "static_websites" {
  value = module.funcs.static_websites
}

output "ws_url" {
  value = module.api.ws_url
}
