terraform {
  required_version = ">= 1.4.0"

  required_providers {
    unarchive = {
      source  = "kota65535/unarchive"
      version = "~> 0.4"
    }
    temporary = {
      source  = "kota65535/temporary"
      version = "~> 1.0"
    }
    shell = {
      source  = "scottwinkler/shell"
      version = "~> 1.7"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}
