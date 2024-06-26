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
  include_files_with_metadata_options = [for e in local.object_metadata : join(" ", [for f in e.filenames : "--include '${f}'"])]
  exclude_files_with_metadata_option  = join(" ", distinct(flatten([for e in local.object_metadata : [for f in e.filenames : "--exclude '${f}'"]])))

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

  modified_files = merge(local.file_replacements, local.json_overrides)
}

data "temporary_directory" "archive" {
  name = "s3-deployment/${md5(var.archive_path)}"
}

data "unarchive_file" "main" {
  type        = "zip"
  source_file = var.archive_path
  patterns    = var.file_patterns
  excludes    = var.file_exclusions
  output_dir  = data.temporary_directory.archive.id
}

// Modify files
data "shell_script" "modifications" {
  for_each = local.modified_files

  lifecycle_commands {
    read = <<-EOT
      cat <<-EOF > '${data.temporary_directory.archive.id}/${each.key}'
      ${each.value[0]}
      EOF
    EOT
  }
  interpreter = ["bash", "-c"]

  depends_on = [data.unarchive_file.main]
}

// As the number of files increases, the output of the `terraform plan` becomes very long and difficult to read.
// So we utilize `aws s3 cp` and `aws s3 sync` to copy almost all objects.
resource "shell_script" "objects" {
  triggers = {
    objects               = filemd5(var.archive_path)
    objects_with_metadata = md5(jsonencode(var.object_metadata))
  }
  // Copy files without metadata
  lifecycle_commands {
    create = <<-EOT
      aws s3 cp --recursive ${data.unarchive_file.main.output_dir} s3://${var.bucket} \
        ${local.exclude_files_with_metadata_option} >&2
    EOT
    read   = <<-EOT
      aws s3api list-objects --bucket ${var.bucket} --query "{Keys:Contents[].Key}" --output json
    EOT
    // If we delete objects when this resource is replaced by changing triggers, there will be a moment when both
    // new and old objects are not present.
    // So we do not perform deletion when the resource is destroyed.
    delete = ""
  }
  interpreter = ["bash", "-c"]

  depends_on = [var.resources_depends_on]
}

// Files with metadata are copied separately
resource "shell_script" "objects_with_metadata" {
  count = length(local.object_metadata)

  triggers = {
    objects               = filemd5(var.archive_path)
    objects_with_metadata = md5(jsonencode(var.object_metadata))
  }
  // Copy files with metadata
  lifecycle_commands {
    create = <<-EOT
      aws s3 cp --recursive ${data.unarchive_file.main.output_dir} s3://${var.bucket} \
        --exclude '*' \
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
    // If we delete objects when this resource is replaced by changing triggers, there will be a moment when both
    // new and old objects are not present.
    // So we do not perform deletion when the resource is destroyed.
    delete = ""
  }
  interpreter = ["bash", "-c"]

  depends_on = [shell_script.objects, var.resources_depends_on]
}

resource "shell_script" "invalidation" {
  triggers = {
    archive_hash      = filemd5(var.archive_path)
    file_replacements = jsonencode(var.file_replacements)
    json_overrides    = jsonencode(var.json_overrides)
    object_metadata   = jsonencode(var.object_metadata)
  }
  // Delete unneeded files & invalidate CloudFront cache
  lifecycle_commands {
    create = <<-EOT
      aws s3 sync --delete ${data.unarchive_file.main.output_dir} s3://${var.bucket}
      if [ "${var.cloudfront_distribution_id}" == "" ]; then
        exit 0
      fi
      invalidation_id=$(aws cloudfront create-invalidation --distribution-id "${var.cloudfront_distribution_id}" --path '/*' --query "Invalidation.Id" --output text)
      while true; do
        sleep 10;
        status=$(aws cloudfront get-invalidation --distribution-id "${var.cloudfront_distribution_id}" --id "$${invalidation_id}" --query "Invalidation.Status" --output text)
        if [[ "$${status}" == "Completed" ]]; then
          break
        fi
      done
    EOT
    read   = <<-EOT
      aws s3api list-objects --bucket ${var.bucket} --query "{Keys:Contents[].Key}" --output json
    EOT
    // Do nothing
    delete = ""
  }
  interpreter = ["bash", "-c"]

  depends_on = [shell_script.objects, shell_script.objects_with_metadata, var.resources_depends_on]
}
