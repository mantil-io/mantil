locals {
  aws_region       = "eu-central-1"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  aws_profile      = "{{.Organization.Name}}"                # TODO profile for use in local aws cli
  dns_zone         = "{{.Organization.DNSZone}}"             # TODO route53 managed zone where dns records will be created
  domain           = "{{.Organization.DNSZone}}"             # TODO api url
  path             = "{{.Name}}"
  project          = "{{.Name}}"
  cert_arn         = "{{.Organization.CertArn}}"             # TODO ssl certificate for the *.domain (created in advance)
  project_bucket   = "{{.Bucket}}"                           # TODO bucket for project configuration/state/functions (created in advance)
  table_name       = "{{.Table.Name}}"
  functions = {
    {{- range .Functions}}
    {{.Name}} = {
      s3_key = "{{.S3Key}}"
      image_key = "{{.ImageKey}}"
      runtime = "{{.Runtime}}"
      public = {{.Public}}
      memory_size = {{.MemorySize}}
      timeout = {{.Timeout}}
      path = "{{.Path}}"
      url = "{{.URL}}"
      handler = "{{.Handler}}"
      env = {
        {{- range $key, $value := .Env}}
        {{$key}} = "{{$value}}"
        {{- end}}
      }
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
}

module "funcs" {
  source        = "http://localhost:8080/terraform/modules/funcs.zip"
  dns_zone      = local.dns_zone
  domain        = local.domain
  api_base_path = local.path
  project       = local.project
  cert_arn      = local.cert_arn
  functions     = local.functions
  s3_bucket     = local.project_bucket
  global_env = {
    domain = local.domain
  }
}

module "dynamodb" {
  source   = "http://localhost:8080/terraform/modules/dynamodb.zip"
  name     = local.table_name
}

# expose aws region and profile for use in shell scripts
output "aws_region" {
  value = local.aws_region
}

output "aws_profile" {
  value = local.aws_profile
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

output "dynamodb_table" {
  value = module.dynamodb.table_name
}
