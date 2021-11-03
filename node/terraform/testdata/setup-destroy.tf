terraform {
  backend "s3" {
    bucket = "bucket-name"
    key    = "setup/terraform/state.tfstate"
    region = "aws-region"
  }
}

provider "aws" {
  region                 = "aws-region"
  skip_get_ec2_platforms = true
}
