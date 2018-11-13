# Terraform Static Registry

This tool can generate a static Terraform module registry. The registry is minimalistic, only providing the endpoints expected by `terraform init`, not the entire registry protocol.

Currently the following server types are supported:

* Caddy
* S3

## Setup

1. Download all of your versioned module files in .tar.gz format (this can be done from the GitHub releases page). 
2. Store all of the downloaded archives in directories by namespace with the proper filename format. This format should match the repository name format expected by the public registry. For example, if you have a module named `test` for the `null` provider, version `2.0.2`, the file name should be `terraform-null-test-2.0.2.tar.gz`.
3. Run the tool against your directory of module, and it will generate the necessary static files. See the specific sections for server types for configuration options.

## Caddy

TBD

## S3

TBD