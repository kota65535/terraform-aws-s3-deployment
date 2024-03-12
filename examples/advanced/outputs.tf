output "s3_objects" {
  value = { for o in module.s3_deployment.s3_objects_modified : o.key => coalesce(o.source_hash, o.content) }
}
