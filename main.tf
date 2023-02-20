locals {
  terraform_tmp_dir  = "${path.root}/.terraform/tmp"
  archive_output_dir = "${local.terraform_tmp_dir}/${filesha256(var.archive_path)}"
  json_overrides = { for e in var.json_overrides : e.filename =>
    jsonencode(merge(
      jsondecode(file("${data.unarchive_file.main.output_dir}/${e.filename}")),
      jsondecode(e.content))
    )
  }
  object_metadata = [
    for e in var.object_metadata : {
      files               = fileset(data.unarchive_file.main.output_dir, e.glob)
      content_type        = e.content_type
      cache_control       = e.cache_control
      content_disposition = e.content_disposition
      content_encoding    = e.content_encoding
      content_language    = e.content_language
    }
  ]
}

data "unarchive_file" "main" {
  type        = "zip"
  source_file = var.archive_path
  pattern     = "**/*"
  output_dir  = local.archive_output_dir
}

resource "aws_s3_bucket_object" "main" {
  for_each = { for i in data.unarchive_file.main.output_files : i.name => i }

  bucket = var.s3_bucket
  key    = each.value.name
  content = (contains(keys(local.json_overrides), each.value.name)
    ? local.json_overrides[each.value.name]
    : file("${data.unarchive_file.main.output_dir}/${each.value.name}")
  )
  content_type = coalescelist(
    [for e in local.object_metadata : e.content_type if contains(e.files, each.value)],
    [try(local.file_types[regex("\\.[^.]+$", each.value)], null)]
  )[0]
  cache_control = coalescelist(
    [for e in local.object_metadata : e.cache_control if contains(e.files, each.value.name)],
    [null]
  )[0]
  content_disposition = coalescelist(
    [for e in local.object_metadata : e.content_disposition if contains(e.files, each.value.name)],
    [null]
  )[0]
  content_encoding = coalescelist(
    [for e in local.object_metadata : e.content_encoding if contains(e.files, each.value.name)],
    [null]
  )[0]
  content_language = coalescelist(
    [for e in local.object_metadata : e.content_language if contains(e.files, each.value.name)],
    [null]
  )[0]
}

resource "null_resource" "invalidation" {
  count = var.cloudfront_distribution_id != null ? 1 : 0
  triggers = {
    archive_sha256  = filesha256(var.archive_path)
    json_overrides  = jsonencode(var.json_overrides)
    object_metadata = jsonencode(var.object_metadata)
  }
  provisioner "local-exec" {
    command = "./${path.module}/scripts/invalidate.sh '${var.cloudfront_distribution_id}'"
  }
}
