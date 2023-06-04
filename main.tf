locals {
  terraform_tmp_dir  = "${path.root}/.terraform/tmp"
  archive_output_dir = "${local.terraform_tmp_dir}/s3-deployment/${coalesce(var.archive_key, var.archive_path)}"
  file_replacements  = { for e in var.file_replacements : sort(fileset(data.unarchive_file.main.output_dir, e.filename))[0] => e.content if length(fileset(data.unarchive_file.main.output_dir, e.filename)) > 0 }
  json_overrides = { for e in var.json_overrides : sort(fileset(data.unarchive_file.main.output_dir, e.filename))[0] =>
    jsonencode(merge(
      jsondecode(file("${data.unarchive_file.main.output_dir}/${sort(fileset(data.unarchive_file.main.output_dir, e.filename))[0]}")),
      jsondecode(e.content))
    ) if length(fileset(data.unarchive_file.main.output_dir, e.filename)) > 0
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

resource "aws_s3_object" "main" {
  for_each = { for e in data.unarchive_file.main.output_files : e.name =>
    e if !contains(keys(local.file_replacements), e.name) && !contains(keys(local.json_overrides), e.name) && !try(length(regex("^META\\-INF(/|\\z)", e.name)) > 0, false)
  }

  bucket      = var.s3_bucket
  key         = each.key
  source      = each.value.path
  source_hash = filemd5(each.value.path)
  content_type = coalescelist(
    compact([for e in local.object_metadata : e.content_type if contains(e.files, each.key)]),
    [try(local.file_types[regex("\\.[^.]+$", each.key)], null)]
  )[0]
  cache_control = coalescelist(
    [for e in local.object_metadata : e.cache_control if contains(e.files, each.key)],
    [null]
  )[0]
  content_disposition = coalescelist(
    [for e in local.object_metadata : e.content_disposition if contains(e.files, each.key)],
    [null]
  )[0]
  content_encoding = coalescelist(
    [for e in local.object_metadata : e.content_encoding if contains(e.files, each.key)],
    [null]
  )[0]
  content_language = coalescelist(
    [for e in local.object_metadata : e.content_language if contains(e.files, each.key)],
    [null]
  )[0]
}

resource "aws_s3_object" "modified" {
  for_each = merge(local.file_replacements, local.json_overrides)

  bucket  = var.s3_bucket
  key     = each.key
  content = each.value
  content_type = coalescelist(
    compact([for e in local.object_metadata : e.content_type if contains(e.files, each.key)]),
    [try(local.file_types[regex("\\.[^.]+$", each.key)], null)]
  )[0]
  cache_control = coalescelist(
    [for e in local.object_metadata : e.cache_control if contains(e.files, each.key)],
    [null]
  )[0]
  content_disposition = coalescelist(
    [for e in local.object_metadata : e.content_disposition if contains(e.files, each.key)],
    [null]
  )[0]
  content_encoding = coalescelist(
    [for e in local.object_metadata : e.content_encoding if contains(e.files, each.key)],
    [null]
  )[0]
  content_language = coalescelist(
    [for e in local.object_metadata : e.content_language if contains(e.files, each.key)],
    [null]
  )[0]
}

moved {
  from = aws_s3_object.json
  to   = aws_s3_object.modified
}

resource "null_resource" "invalidation" {
  count = var.cloudfront_distribution_id != null ? 1 : 0
  triggers = {
    archive_hash      = filemd5(var.archive_path)
    file_replacements = jsonencode(var.file_replacements)
    json_overrides    = jsonencode(var.json_overrides)
    object_metadata   = jsonencode(var.object_metadata)
  }
  provisioner "local-exec" {
    command = "./${path.module}/scripts/invalidate.sh '${var.cloudfront_distribution_id}'"
  }
}
