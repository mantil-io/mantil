variable "functions" {
  default     = {}
  description = "Definition of lambda functions. See main.tf locals for attributes."
}

variable "s3_bucket" {
  type        = string
  default     = null
  description = "S3 bucket containing functions' deployment package."
}

variable "prefix" {
  type = string
}

variable "suffix" {
  type = string
}
