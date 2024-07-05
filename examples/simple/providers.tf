terraform {
  backend "s3" {
    bucket = "terraform-backend-561678142736"
    region = "ap-northeast-1"
    key    = "terraform-aws-s3-deployment-simple.tfstate"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.67.0"
    }
    temporary = {
      source  = "kota65535/temporary"
      version = "0.2.1"
    }
    unarchive = {
      source  = "kota65535/unarchive"
      version = "0.4.1"
    }
    shell = {
      source  = "scottwinkler/shell"
      version = "1.7.10"
    }
  }
  required_version = ">= 1.4.0"
}

provider "aws" {
  region = "ap-northeast-1"
}

provider "temporary" {
  base = "${path.root}/.terraform/tmp"
}
