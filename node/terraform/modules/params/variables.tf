variable "path_prefix" {
  type = string
}

variable "params" {
  type = list(object({
    name   = string
    value  = string
    secure = optional(bool)
  }))
  default = []
}
