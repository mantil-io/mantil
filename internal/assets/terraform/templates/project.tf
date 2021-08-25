locals {
  aws_region       = "eu-central-1"                          # TODO region where resources will be created (except cloudfront distribution which is global)
  aws_profile      = "{{.Organization.Name}}"                # TODO profile for use in local aws cli
  dns_zone         = "{{.Organization.DNSZone}}"             # TODO route53 managed zone where dns records will be created
  domain           = "{{.Organization.DNSZone}}"             # TODO api url
  path             = "{{.Name}}"
  project          = "{{.Name}}"
  cert_arn         = "{{.Organization.CertArn}}"             # TODO ssl certificate for the *.domain (created in advance)
  project_bucket   = "{{.Bucket}}"                           # TODO bucket for project configuration/state/functions (created in advance)
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
  static_websites = local.static_websites
  global_env = {
    domain = local.domain
  }
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
