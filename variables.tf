variable "archive_path" {
  description = "Path of the archive file containing your static website resources"
  type        = string
}

variable "bucket" {
  description = "Name of a S3 bucket for hosting your static website"
  type        = string
}

variable "cloudfront_distribution_id" {
  description = "CloudFront distribution ID. Used to invalidate the cache when any resources has changed"
  type        = string
  default     = null
}

variable "file_patterns" {
  description = "[Glob patterns](https://developer.hashicorp.com/terraform/language/functions/fileset) to filter files when extracting the archive"
  type        = list(string)
  default     = null
}

variable "file_exclusions" {
  description = "[Glob patterns](https://developer.hashicorp.com/terraform/language/functions/fileset) to exclude files when extracting the archive"
  type        = list(string)
  default     = null
}

variable "file_replacements" {
  description = <<-EOT
  File replacement settings.
  * filename : Name of the file to be replaced. [Glob pattern](https://developer.hashicorp.com/terraform/language/functions/fileset) is available. If patterns of the multiple settings match, only the first matched one is used.
  * content  : Content string to store in the file
EOT
  type = list(object({
    filename = string
    content  = string
  }))
  default = []
}

variable "json_overrides" {
  description = <<-EOT
  JSON override settings.
  * filename : Name of a JSON file whose properties will be overridden. [Glob pattern](https://developer.hashicorp.com/terraform/language/functions/fileset) is available. If patterns of the multiple settings match, only the first matched one is used.
  * content  : JSON string whose properties will override them
EOT
  type = list(object({
    filename = string
    content  = string
  }))
  default = []
}

variable "object_metadata" {
  description = <<-EOT
  Object metadata settings.
  * glob                : [Glob pattern](https://developer.hashicorp.com/terraform/language/functions/fileset) to match files to set metadata values. If patterns of the multiple settings match, only the first matched one is used.
  * cache_control       : Cache-Control metadata value
  * content_disposition : Content-Disposition metadata value
  * content_encoding    : Content-Encoding metadata value
  * content_language    : Content-Language metadata value
  * content_type        : Content-Type metadata value
EOT
  type = list(object({
    glob                = string
    cache_control       = optional(string)
    content_disposition = optional(string)
    content_encoding    = optional(string)
    content_language    = optional(string)
    content_type        = optional(string)
  }))
  default = []
}

variable "resources_depends_on" {
  description = "Optional 'depends_on' values for resources only to control the deployment order"
  type        = list(any)
  default     = []
}
