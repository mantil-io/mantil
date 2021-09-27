terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "{{.BucketPrefix}}terraform/state.tfstate"
    region = "{{.Region}}"
  }
}

provider "aws" {
  region                 = "{{.Region}}"
  skip_get_ec2_platforms = true
}
