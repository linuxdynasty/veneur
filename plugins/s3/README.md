S3 Plugin
===========

The S3 plugin archives every flush to S3 as a separate S3 object.

This plugin is still in an experimental state.


# Config Options based on format

The following options are available to use with the S3 plugin.

* plugins_output: `string`
  * Available options as of today = `tsv`, `tsdb`, or `wavefront`
  * default = `tsv` extension `.tsv`
  * `wavefront` and `tsdb` does not have an extension.
* plugins_output_compressed: `bool`
  * default = `true`
  * if `true` it will compress and add the `.gz` extension.
* plugins_output_file_name_structure: `string`
  * Available options as of today = `date_host` or `""`
  * default = `""`
  * `""` means to store in the root of the s3 bucket with no prefix.
  * `"date_host"` means to add the `year/month/day/veneur_host/` as the prefix to the metric name.
* plugins_output_name_type: `string`
  * Available options as of today = `uuid` or `timestamp`
  * default = `timestamp`

# Config Options to connect to S3

The only 2 options that are mandatory are.

* aws_s3_bucket: `string`
* aws_region: `string`

These parameters are optional.

* aws_access_key_id `string`
* aws_secret_access_key `string`

The Golang AWS SDK will load up Credentials in the following order. https://docs.aws.amazon.com/sdk-for-go/api/aws/session/

1. Environment Variables
2. Shared Credentials file
3. Shared Configuration file (if SharedConfig is enabled) `export AWS_SDK_LOAD_CONFIG=1`
4. EC2 Instance Metadata (credentials only)