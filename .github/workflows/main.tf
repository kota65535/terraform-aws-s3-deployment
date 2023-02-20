locals {
  bucket_name = "s3-deployment-561678142736"
}

module "s3_deployment" {
  source = "../../"

  archive_path = "test.zip"
  s3_bucket    = local.bucket_name
  json_modifications = [
    {
      filename = "b.json"
      content = {
        foo = "bar"
      }
    }
  ]
  object_metadata = [
    {
      glob          = "*.html"
      cache_control = "public, max-age=0, must-revalidate"
    },
    {
      glob          = "*.js"
      cache_control = "public, max-age=0, must-revalidate"
    }
  ]
  cloudfront_distribution_id = aws_cloudfront_distribution.main.id
}

resource "aws_s3_bucket" "main" {
  bucket = local.bucket_name

  cors_rule {
    allowed_headers = [
      "Authorization",
      "Content-Length"
    ]
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
    max_age_seconds = 3000
  }

  versioning {
    enabled = true
  }
}

#output "objects" {
#  value = module.s3_deployment.objects
#}
