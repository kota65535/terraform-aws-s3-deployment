module "s3_deployment" {
  source = "../../"

  archive_path = "test.zip"
  bucket       = aws_s3_bucket.main.bucket
}
