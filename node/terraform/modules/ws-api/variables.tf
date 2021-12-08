variable "suffix" {
  type = string
}

variable "naming_template" {
  type = string
}

variable "functions_bucket" {
  type = string
}

variable "functions_s3_path" {
  type = string
}

variable "ws_env" {
  type    = map(any)
  default = {}
}

variable "authorizer" {
  type = object({
    authorization_header = string
    arn                  = string
    invoke_arn           = string
  })
  default = null
}

variable "domain" {
  type    = string
  default = ""
}
