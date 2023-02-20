output "objects" {
  value = merge(aws_s3_object.main, aws_s3_object.json)
}
