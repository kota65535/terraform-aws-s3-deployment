output "s3_objects" {
  value       = merge(aws_s3_object.main, aws_s3_object.json)
  description = "Uploaded S3 objects"
}
