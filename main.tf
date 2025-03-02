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
  files_with_metadata = distinct(flatten([for e in local.object_metadata : e.filenames]))

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

  aws_config_environments = <<-EOT
    %{~if var.aws_config.access_key != null~}
    export AWS_ACCESS_KEY_ID='${var.aws_config.access_key}'
    %{~endif~}
    %{~if var.aws_config.secret_key != null~}
    export AWS_SECRET_ACCESS_KEY='${var.aws_config.secret_key}'
    %{~endif~}
    %{~if var.aws_config.region != null~}
    export AWS_REGION='${var.aws_config.region}'
    %{~endif~}
    %{~if var.aws_config.profile != null~}
    export AWS_PROFILE='${var.aws_config.profile}'
    %{~endif~}
  EOT
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
// So we utilize `aws s3 cp` and `aws s3 sync` to copy all objects.
resource "shell_script" "objects" {
  triggers = {
    archive_hash      = filemd5(var.archive_path)
    bucket            = var.bucket
    file_patterns     = jsonencode(var.file_patterns)
    file_exclusions   = jsonencode(var.file_exclusions)
    file_replacements = jsonencode(var.file_replacements)
    json_overrides    = jsonencode(var.json_overrides)
    object_metadata   = jsonencode(var.object_metadata)
  }
  lifecycle_commands {
    // The create command does the following:
    // 1. Copy objects without metadata using `aws s3 cp`.
    // 2. Copy objects with metadata. This is done for each object_metadata setting.
    //    If a file matches with glob patterns of the multiple settings entries, only the first matched one is used
    // 3. Delete unneeded objects using `aws s3 sync --delete`.
    create = <<-EOT
      set -eEuo pipefail
      export LC_ALL=C

      ${local.aws_config_environments}

      cd ${data.unarchive_file.main.output_dir}
      aws s3 cp --recursive . s3://${var.bucket} ${join(" ", [for f in local.files_with_metadata : "--exclude '${f}'"])} >&2
      %{~for i, om in reverse(local.object_metadata)~}
      aws s3 cp --recursive . s3://${var.bucket} \
        --exclude '*' ${join(" ", [for f in om.filenames : "--include '${f}'"])} \
        %{~if om.content_type != null~}
        --content-type '${om.content_type}' \
        %{~endif~}
        %{~if om.cache_control != null~}
        --cache-control '${om.cache_control}' \
        %{~endif~}
        %{~if om.content_disposition != null~}
        --content-disposition '${om.content_disposition}' \
        %{~endif~}
        %{~if om.content_encoding != null~}
        --content-encoding '${om.content_encoding}' \
        %{~endif~}
        %{~if om.content_language != null~}
        --content-language '${om.content_language}' \
        %{~endif~}
        --metadata-directive REPLACE >&2
      %{~endfor~}
      aws s3 sync --delete . s3://${var.bucket}
    EOT
    read   = <<-EOT
      set -eEuo pipefail
      export LC_ALL=C

      ${local.aws_config_environments}

      aws s3api list-objects --bucket ${var.bucket} --query "{Keys:Contents[].Key}" --output json
    EOT
    // If we delete objects when this resource is replaced by changing triggers, there will be a moment when both
    // new and old objects are not present.
    // So we do not perform deletion when the resource is destroyed.
    delete = ""
  }
  interpreter = ["bash", "-c"]

  depends_on = [data.shell_script.modifications, var.resources_depends_on]
  lifecycle {
    ignore_changes = [lifecycle_commands]
  }
}

// Invalidate CloudFront cache
resource "shell_script" "invalidation" {
  triggers = {
    archive_hash      = filemd5(var.archive_path)
    file_patterns     = jsonencode(var.file_patterns)
    file_exclusions   = jsonencode(var.file_exclusions)
    file_replacements = jsonencode(var.file_replacements)
    json_overrides    = jsonencode(var.json_overrides)
    object_metadata   = jsonencode(var.object_metadata)
  }
  // Create CloudFront invalidation and wait until completion.
  lifecycle_commands {
    create = <<-EOT
      set -eEuo pipefail
      export LC_ALL=C

      ${local.aws_config_environments}

      if [ -z "${var.cloudfront_distribution_id}" ]; then
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
      set -eEuo pipefail
      export LC_ALL=C

      ${local.aws_config_environments}

      aws s3api list-objects --bucket ${var.bucket} --query "{Keys:Contents[].Key}" --output json
    EOT
    // Do nothing
    delete = ""
  }
  interpreter = ["bash", "-c"]

  depends_on = [shell_script.objects]
}
