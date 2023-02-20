terraform {
  backend "s3" {
    bucket = "terraform-backend-561678142736"
    region = "ap-northeast-1"
    key    = "terraform-aws-s3-deployment.tfstate"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.48.0"
    }
  }
  required_version = "~> 1.3.0"
}

provider "aws" {
  region = "ap-northeast-1"
}
