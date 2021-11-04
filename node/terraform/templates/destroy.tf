terraform {
  backend "s3" {
    bucket = "{{.Bucket}}"
    key    = "{{.BucketPrefix}}/state.tfstate"
    region = "{{.Region}}"
  }
}

provider "aws" {
  region                 = "{{.Region}}"
  skip_get_ec2_platforms = true
}
