variable "functions" {
  default     = {}
  description = "Definition of lambda functions. See main.tf locals for attributes."
}

variable "global_env" {
  type        = map(any)
  default     = {}
  description = "Global environment variables for all functions."
}

variable "api_stage_name" {
  type        = string
  default     = "main"
  description = "Api gateway stage name."
}

variable "project_name" {
  type        = string
  description = "Name of the project for which functions are being created."
}

variable "s3_bucket" {
  type        = string
  default     = null
  description = "S3 bucket containing functions' deployment package."
}

variable "static_websites" {
  default = {}
}
