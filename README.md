# terraform-aws-s3-deployment

Terraform module which deploys a static website to a S3 bucket.

See: https://registry.terraform.io/modules/kota65535/s3-deployment/aws/

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.4.0 |
| <a name="requirement_shell"></a> [shell](#requirement\_shell) | ~> 1.7 |
| <a name="requirement_temporary"></a> [temporary](#requirement\_temporary) | ~> 0.2 |
| <a name="requirement_unarchive"></a> [unarchive](#requirement\_unarchive) | ~> 0.4 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_shell"></a> [shell](#provider\_shell) | ~> 1.7 |
| <a name="provider_temporary"></a> [temporary](#provider\_temporary) | ~> 0.2 |
| <a name="provider_unarchive"></a> [unarchive](#provider\_unarchive) | ~> 0.4 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [shell_script.invalidation](https://registry.terraform.io/providers/scottwinkler/shell/latest/docs/resources/script) | resource |
| [shell_script.objects](https://registry.terraform.io/providers/scottwinkler/shell/latest/docs/resources/script) | resource |
| [shell_script.objects_with_metadata](https://registry.terraform.io/providers/scottwinkler/shell/latest/docs/resources/script) | resource |
| [shell_script.modifications](https://registry.terraform.io/providers/scottwinkler/shell/latest/docs/data-sources/script) | data source |
| [temporary_directory.archive](https://registry.terraform.io/providers/kota65535/temporary/latest/docs/data-sources/directory) | data source |
| [unarchive_file.main](https://registry.terraform.io/providers/kota65535/unarchive/latest/docs/data-sources/file) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_archive_path"></a> [archive\_path](#input\_archive\_path) | Path of the archive file containing your static website resources | `string` | n/a | yes |
| <a name="input_bucket"></a> [bucket](#input\_bucket) | Name of a S3 bucket for hosting your static website | `string` | n/a | yes |
| <a name="input_cloudfront_distribution_id"></a> [cloudfront\_distribution\_id](#input\_cloudfront\_distribution\_id) | CloudFront distribution ID. Used to invalidate the cache when any resources has changed | `string` | `""` | no |
| <a name="input_file_exclusions"></a> [file\_exclusions](#input\_file\_exclusions) | [Glob patterns](https://developer.hashicorp.com/terraform/language/functions/fileset) to exclude files when extracting the archive | `list(string)` | `null` | no |
| <a name="input_file_patterns"></a> [file\_patterns](#input\_file\_patterns) | [Glob patterns](https://developer.hashicorp.com/terraform/language/functions/fileset) to filter files when extracting the archive | `list(string)` | `null` | no |
| <a name="input_file_replacements"></a> [file\_replacements](#input\_file\_replacements) | File replacement settings.<br>* filename : Name of the file to be replaced. [Glob pattern](https://developer.hashicorp.com/terraform/language/functions/fileset) is available. If patterns of the multiple settings match, only the first matched one is used.<br>* content  : Content string to store in the file | <pre>list(object({<br>    filename = string<br>    content  = string<br>  }))</pre> | `[]` | no |
| <a name="input_json_overrides"></a> [json\_overrides](#input\_json\_overrides) | JSON override settings.<br>* filename : Name of a JSON file whose properties will be overridden. [Glob pattern](https://developer.hashicorp.com/terraform/language/functions/fileset) is available. If patterns of the multiple settings match, only the first matched one is used.<br>* content  : JSON string whose properties will override them | <pre>list(object({<br>    filename = string<br>    content  = string<br>  }))</pre> | `[]` | no |
| <a name="input_object_metadata"></a> [object\_metadata](#input\_object\_metadata) | Object metadata settings.<br>* glob                : Glob pattern to match files to set metadata values. The patterns must not be overlapped.<br>* cache\_control       : Cache-Control metadata value<br>* content\_disposition : Content-Disposition metadata value<br>* content\_encoding    : Content-Encoding metadata value<br>* content\_language    : Content-Language metadata value<br>* content\_type        : Content-Type metadata value | <pre>list(object({<br>    glob                = string<br>    cache_control       = optional(string)<br>    content_disposition = optional(string)<br>    content_encoding    = optional(string)<br>    content_language    = optional(string)<br>    content_type        = optional(string)<br>  }))</pre> | `[]` | no |
| <a name="input_resources_depends_on"></a> [resources\_depends\_on](#input\_resources\_depends\_on) | Optional 'depends\_on' values for resources only to control the deployment order | `list(any)` | `[]` | no |

## Outputs

No outputs.
<!-- END_TF_DOCS -->
