# terraform-aws-s3-deployment

Terraform module which deploys a static website to a S3 bucket.

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_unarchive"></a> [unarchive](#requirement\_unarchive) | ~> 0.3.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | n/a |
| <a name="provider_null"></a> [null](#provider\_null) | n/a |
| <a name="provider_unarchive"></a> [unarchive](#provider\_unarchive) | ~> 0.3.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_s3_object.json](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_object) | resource |
| [aws_s3_object.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_object) | resource |
| [null_resource.invalidation](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource) | resource |
| [unarchive_file.main](https://registry.terraform.io/providers/kota65535/unarchive/latest/docs/data-sources/file) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_archive_path"></a> [archive\_path](#input\_archive\_path) | Path of an archive file containing your static website resources | `string` | n/a | yes |
| <a name="input_cloudfront_distribution_id"></a> [cloudfront\_distribution\_id](#input\_cloudfront\_distribution\_id) | CloudFront distribution ID. Used to invalidate the cache when any resources has changed | `string` | `null` | no |
| <a name="input_json_overrides"></a> [json\_overrides](#input\_json\_overrides) | JSON override settings.<br>filename : Name of a JSON file whose properties are to be overridden<br>content  : JSON string whose properties will override them | <pre>list(object({<br>    filename = string<br>    content  = string<br>  }))</pre> | `[]` | no |
| <a name="input_object_metadata"></a> [object\_metadata](#input\_object\_metadata) | Object metadata settings.<br>glob                : Glob pattern to match files to set metadata values<br>content\_type        : Content-Type metadata value<br>cache\_control       : Cache-Control metadata value<br>content\_disposition : Content-Disposition metadata value<br>content\_encoding    : Content-Encoding metadata value<br>content\_language    : Content-Language metadata value | <pre>list(object({<br>    glob                = string<br>    content_type        = optional(string)<br>    cache_control       = optional(string)<br>    content_disposition = optional(string)<br>    content_encoding    = optional(string)<br>    content_language    = optional(string)<br>  }))</pre> | `[]` | no |
| <a name="input_s3_bucket"></a> [s3\_bucket](#input\_s3\_bucket) | Name of a S3 bucket for hosting your static website | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_s3_objects"></a> [s3\_objects](#output\_s3\_objects) | Uploaded S3 objects |
