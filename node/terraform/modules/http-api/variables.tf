variable "prefix" {
  type = string
}

variable "suffix" {
  type = string
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

variable "authorizer" {
  type = object({
    authorization_header = string
    arn                  = string
    invoke_arn           = string
  })
  default = null
}
