locals {
  aws_region       = "eu-central-1"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  project_name     = "{{.Name}}"
  project_bucket   = "{{.Bucket}}"                           # TODO bucket for project configuration/state/functions (created in advance)
  functions_bucket = "mantil-downloads"
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
    {{- range .StaticWebsites}}
    {{.Name}} = {
      name = "{{.Name}}"
    }
    {{- end}}
  }
  ws_handler = {
    name        = "ws-handler"
    s3_key      = "functions/ws-handler.zip"
    memory_size = 128
    timeout     = 900
  }
  ws_sqs_forwarder = {
    name        = "ws-sqs-forwarder"
    s3_key      = "functions/ws-sqs-forwarder.zip"
    memory_size = 128
    timeout     = 900
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
  source          = "http://localhost:8080/terraform/modules/funcs.zip"
  project_name    = local.project_name
  functions       = local.functions
  s3_bucket       = local.project_bucket
  static_websites = local.static_websites
}

module "ws" {
  source        = "http://localhost:8080/terraform/modules/ws.zip"
  handler       = local.ws_handler
  sqs_forwarder = local.ws_sqs_forwarder
  s3_bucket     = local.functions_bucket
  project_name  = local.project_name
}

output "url" {
  value = module.funcs.url
}

output "functions" {
  value = module.funcs.functions
}

output "functions_bucket" {
  value = local.project_bucket
}

output "static_websites" {
  value = module.funcs.static_websites
}

output "ws_url" {
  value = module.ws.url
}

output "ws_handler" {
  value = module.ws.handler
}

output "ws_sqs_forwarder" {
  value = module.ws.sqs_forwarder
}
