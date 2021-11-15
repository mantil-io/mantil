variable "prefix" {
  type = string
}

variable "suffix" {
  type = string
}

variable "ws_enabled" {
  type    = bool
  default = false
}

variable "integrations" {
  type = list(object({
    type               = string
    method             = string
    integration_method = string
    route              = string
    uri                = string
    lambda_name        = optional(string)
    enable_auth        = optional(bool)
    is_default         = optional(bool)
  }))
  default = []
}

variable "functions_bucket" {
  type = string
}

variable "functions_s3_path" {
  type = string
}

variable "authorizer" {
  type = object({
    authorization_header = string
    env                  = map(string)
  })
  default = null
}

variable "ws_env" {
  type    = map(any)
  default = {}
}
