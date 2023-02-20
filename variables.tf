variable "archive_path" {
  description = "Path of an archive file to extract and deploy"
  type        = string
}

variable "s3_bucket" {
  description = "S3 bucket name for hosting"
  type        = string
}

variable "cloudfront_distribution_id" {
  description = "CloudFront distribution ID to invalidate cache"
  type        = string
  default     = null
}

variable "json_modifications" {
  description = "JSON files to modify"
  type = list(object({
    filename = string
    content  = any
  }))
  default = []
}

variable "object_metadata" {
  description = "Object metadata specification"
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
