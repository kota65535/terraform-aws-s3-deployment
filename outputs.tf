output "objects" {
  value = merge(aws_s3_bucket_object.main, aws_s3_bucket_object.json)
}
