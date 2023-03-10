variable "archive_path" {
  description = "Path of an archive file containing your static website resources"
  type        = string
}

variable "s3_bucket" {
  description = "Name of a S3 bucket for hosting your static website"
  type        = string
}

variable "cloudfront_distribution_id" {
  description = "CloudFront distribution ID. Used to invalidate the cache when any resources has changed"
  type        = string
  default     = null
}

variable "json_overrides" {
  description = <<-EOT
  JSON override settings.
  filename : Name of a JSON file whose properties are to be overridden
  content  : JSON string whose properties will override them
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
  glob                : Glob pattern to match files to set metadata values
  content_type        : Content-Type metadata value
  cache_control       : Cache-Control metadata value
  content_disposition : Content-Disposition metadata value
  content_encoding    : Content-Encoding metadata value
  content_language    : Content-Language metadata value
EOT
  type = list(object({
    glob                = string
    content_type        = optional(string)
    cache_control       = optional(string)
    content_disposition = optional(string)
    content_encoding    = optional(string)
    content_language    = optional(string)
  }))
  default = []
}
