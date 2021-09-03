locals {
  aws_region     = "eu-central-1"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  project_name   = "{{.Name}}"
  project_bucket = "{{.Bucket}}"                           # TODO bucket for project configuration/state/functions (created in advance)
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
