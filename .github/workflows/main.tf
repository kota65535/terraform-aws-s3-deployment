locals {
  bucket_name = "s3-deployment-561678142736"
  config = {
    unicorns = "awesome"
    abc      = [1, 2, 3]
  }
}

module "s3_deployment" {
  source = "../../"

  archive_path = "test.zip"
  s3_bucket    = local.bucket_name
  file_replacements = [
    {
      filename = "config-*.js"
      content  = "const c = JSON.parse('${jsonencode(local.config)}'); export default c;"
    }
  ]
  json_overrides = [
    {
      filename = "b.json"
      content = jsonencode({
        h = "2"
        i = {
          j = 3
          k = "4"
        }
      })
    }
  ]
  object_metadata = [
    {
      glob                = "b.json"
      cache_control       = "public, max-age=31536000, immutable"
      content_disposition = "inline"
      content_encoding    = "compress"
      content_language    = "ja-JP"
    },
    {
      glob             = "*.json"
      content_language = "en-US"
    },
    {
      glob          = "*.html"
      cache_control = "public, max-age=0, must-revalidate"
    },
    {
      glob          = "*.js"
      content_type  = "text/javascript"
      cache_control = "public, max-age=0, must-revalidate"
    }
  ]
  cloudfront_distribution_id = aws_cloudfront_distribution.main.id
}

resource "aws_s3_bucket" "main" {
  bucket = local.bucket_name
}

resource "aws_s3_bucket_policy" "main" {
  bucket = aws_s3_bucket.main.bucket
  policy = data.aws_iam_policy_document.oai.json
}

output "s3_objects" {
  value = { for o in module.s3_deployment.s3_objects : o.key => coalesce(o.source_hash, o.content) }
}
