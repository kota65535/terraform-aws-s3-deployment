locals {
  object_metadata = [
    for e in var.object_metadata : {
      filenames           = fileset(data.unarchive_file.main.output_dir, e.glob)
      content_type        = e.content_type
      cache_control       = e.cache_control
      content_disposition = e.content_disposition
      content_encoding    = e.content_encoding
      content_language    = e.content_language
    }
  ]
  include_files_with_metadata_options = [for e in local.object_metadata : join(" ", [for f in setsubtract(e.filenames, keys(local.modified_files)) : "--include '${f}'"])]
  exclude_files_with_metadata_option  = join(" ", distinct(flatten([for e in local.object_metadata : [for f in e.filenames : "--exclude '${f}'"]])))

  // Mapping of filename to metadata.
  // If a file matches with glob patterns of the multiple settings entries, only the first matched one is used
  object_metadata_map = {
    for d in flatten([
      for e in local.object_metadata : [
        for f in e.filenames : {
          filename            = f
          content_type        = e.content_type
          cache_control       = e.cache_control
          content_disposition = e.content_disposition
          content_encoding    = e.content_encoding
          content_language    = e.content_language
      }]
    ]) : d.filename => d...
  }

  // Mapping of filename to the replaced content.
  // If a file matches with glob patterns of the multiple settings entries, only the first matched one is used
  file_replacements = {
    for d in flatten([
      for e in var.file_replacements : [
        for f in fileset(data.unarchive_file.main.output_dir, e.filename) : {
          filename : f
          content : e.content
      }]
    ]) : d.filename => d.content...
  }

  // Mapping of JSON filename to the merged content.
  // If a file matches with glob patterns of the multiple settings entries, only the first matched one is used
  json_overrides = {
    for d in flatten([
      for e in var.json_overrides : [
        for f in fileset(data.unarchive_file.main.output_dir, e.filename) : {
          filename : f
          content : jsonencode(merge(jsondecode(file("${data.unarchive_file.main.output_dir}/${f}")), jsondecode(e.content)))
        }
      ]
    ]) : d.filename => d.content...
  }

  modified_files                = merge(local.file_replacements, local.json_overrides)
  exclude_modified_files_option = join(" ", distinct([for k, v in local.modified_files : "--exclude '${k}'"]))
}

// As the number of files increases, the output of the `terraform plan` becomes very long and difficult to read.
// So we utilize `aws s3 cp` and `aws s3 sync` to copy almost all objects.
resource "shell_script" "objects" {
  triggers = {
    objects               = filemd5(var.archive_path),
    objects_with_metadata = md5(jsonencode(var.object_metadata))
  }
  // Copy all files except those to be modified or those with metadata 
  lifecycle_commands {
    create = <<-EOT
      aws s3 sync --delete ${data.unarchive_file.main.output_dir} s3://${var.bucket} \
        ${local.exclude_files_with_metadata_option} \
        ${local.exclude_modified_files_option} >&2
    EOT
    read   = <<-EOT
      aws s3api list-objects --bucket ${var.bucket} --query "{Keys:Contents[].Key}" --output json
    EOT
    update = <<-EOT
      aws s3 sync --delete ${data.unarchive_file.main.output_dir} s3://${var.bucket} \
        ${local.exclude_files_with_metadata_option} \
        ${local.exclude_modified_files_option} >&2
    EOT
    delete = <<-EOT
      aws s3 rm --recursive s3://${var.bucket} \
        ${local.exclude_files_with_metadata_option} \
        ${local.exclude_modified_files_option} >&2
    EOT
  }
  interpreter = ["bash", "-c"]
}

// Files with metadata are copied separately
resource "shell_script" "objects_with_metadata" {
  count = length(local.object_metadata)
  triggers = {
    objects               = filemd5(var.archive_path),
    objects_with_metadata = md5(jsonencode(var.object_metadata))
  }
  // Copy files with metadata, excluding those to be modified
  lifecycle_commands {
    create = <<-EOT
      aws s3 sync --delete ${data.unarchive_file.main.output_dir} s3://${var.bucket} \
        --exclude "*" \
        ${local.include_files_with_metadata_options[count.index]} \
        %{~if local.object_metadata[count.index].content_type != null~}
        --content-type '${local.object_metadata[count.index].content_type}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].cache_control != null~}
        --cache-control '${local.object_metadata[count.index].cache_control}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].content_disposition != null~}
        --content-disposition '${local.object_metadata[count.index].content_disposition}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].content_encoding != null~}
        --content-encoding '${local.object_metadata[count.index].content_encoding}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].content_language != null~}
        --content-language '${local.object_metadata[count.index].content_language}' \
        %{~endif~}
        --metadata-directive REPLACE >&2
    EOT
    read   = <<-EOT
      aws s3api list-objects --bucket ${var.bucket} --query "{Keys:Contents[].Key}" --output json
    EOT
    update = <<-EOT
      aws s3 sync --delete ${data.unarchive_file.main.output_dir} s3://${var.bucket} \
        --exclude "*" \
        ${local.include_files_with_metadata_options[count.index]} \
        %{~if local.object_metadata[count.index].content_type != null~}
        --content-type '${local.object_metadata[count.index].content_type}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].cache_control != null~}
        --cache-control '${local.object_metadata[count.index].cache_control}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].content_disposition != null~}
        --content-disposition '${local.object_metadata[count.index].content_disposition}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].content_encoding != null~}
        --content-encoding '${local.object_metadata[count.index].content_encoding}' \
        %{~endif~}
        %{~if local.object_metadata[count.index].content_language != null~}
        --content-language '${local.object_metadata[count.index].content_language}' \
        %{~endif~}
        --metadata-directive REPLACE >&2
    EOT
    delete = <<-EOT
      aws s3 rm --recursive s3://${var.bucket} \
        --exclude "*" \
        ${local.include_files_with_metadata_options[count.index]} \
        ${local.exclude_modified_files_option} >&2
    EOT
  }
  interpreter = ["bash", "-c"]
  depends_on  = [shell_script.objects]
}

// Use `aws_s3_object` resource for modified files
resource "aws_s3_object" "modified" {
  for_each = local.modified_files

  bucket              = var.bucket
  key                 = each.key
  content             = each.value[0]
  content_type        = try(local.object_metadata_map[each.key][0].content_type, try(local.file_types[regex("\\.[^.]+$", each.key)], null))
  cache_control       = try(local.object_metadata_map[each.key][0].cache_control, null)
  content_disposition = try(local.object_metadata_map[each.key][0].content_disposition, null)
  content_encoding    = try(local.object_metadata_map[each.key][0].content_encoding, null)
  content_language    = try(local.object_metadata_map[each.key][0].content_language, null)

  depends_on = [shell_script.objects, shell_script.objects_with_metadata]
}

resource "terraform_data" "invalidation" {
  triggers_replace = {
    archive_hash      = filemd5(var.archive_path)
    file_replacements = jsonencode(var.file_replacements)
    json_overrides    = jsonencode(var.json_overrides)
    object_metadata   = jsonencode(var.object_metadata)
  }
  provisioner "local-exec" {
    command     = "./${path.module}/scripts/invalidate.sh '${var.cloudfront_distribution.id}'"
    interpreter = ["bash", "-c"]
  }
  depends_on = [shell_script.objects, shell_script.objects_with_metadata, aws_s3_object.modified]
}

data "temporary_directory" "archive" {
  name = "s3-deployment/${md5(var.archive_path)}"
}

data "unarchive_file" "main" {
  type        = "zip"
  source_file = var.archive_path
  patterns    = var.file_patterns
  excludes    = var.file_exclusion
  output_dir  = data.temporary_directory.archive.id
}
