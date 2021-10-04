variable "name_prefix" {
  type = string
}

variable "integrations" {
  type = list(object({
    type        = string
    method      = string
    route       = string
    uri         = string
    lambda_name = optional(string)
    enable_auth = optional(bool)
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
    public_key           = string
    authorization_header = string
  })
  default = null
}

variable "project_name" {
  type = string
  default = ""
}
