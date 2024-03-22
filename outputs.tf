output "s3_objects_modified" {
  value       = aws_s3_object.modified
  description = "S3 objects replaced or overridden"
}
