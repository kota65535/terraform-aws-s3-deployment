resource "aws_s3_bucket" "main" {
  bucket = local.bucket_name
}
