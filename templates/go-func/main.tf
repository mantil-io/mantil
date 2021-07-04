terraform {
  backend "s3" {
    bucket = "atoz-technology-terraform-state"
    key    = "{{.Organization.Name}}/{{.Name}}.tfstate"
    region = "eu-central-1"
  }
}

locals {
  aws_region       = "eu-central-1"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  aws_profile      = "{{.Organization.Name}}"                # TODO profile for use in local aws cli
  dns_zone         = "{{.Organization.DNSZone}}"             # TODO route53 managed zone where dns records will be created
  domain           = "{{.ApiURL}}"                           # TODO api url
  cert_arn         = "{{.Organization.CertArn}}"             # TODO ssl certificate for the *.domain (created in advance)
  functions_bucket = "{{.Organization.FunctionsBucket}}"     # TODO bucket where lambda functions are deployed (created in advance)
  functions = {
    {{- range .Functions}}
    {{.Name}} = {
      s3_key = "{{.S3Key}}"
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

provider "aws" {
  region                  = local.aws_region
}

module "funcs" {
  source     = "./.modules/terraform-aws-modules/funcs"
  dns_zone   = local.dns_zone
  domain     = local.domain
  cert_arn   = local.cert_arn
  functions  = local.functions
  s3_bucket  = local.functions_bucket
  global_env = {
    domain = local.domain
  }
}

# expose aws region and profile for use in shell scripts
output "aws_region" {
  value = local.aws_region
}

output "aws_profile" {
  value = local.aws_profile
}

output "api_url" {
  value = "https://${local.domain}"
}

output "functions" {
  value = module.funcs.functions
}
output "functions_bucket" {
  value = local.functions_bucket
}

