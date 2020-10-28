# Terraform Static Registry

This tool can generate a static Terraform Provider registry. The registry is minimalistic, only providing the endpoints expected by `terraform init`, not the entire registry protocol.

This tool supports multiple different sources of provider information and multiple static server types.

To get started you create a configuration file similar to the following to tell the tool where to find your provider information:

```hcl
provider "paultyng" "unifi" {
  github {
    repository = "paultyng/terraform-provider-unifi"
    public_key_file = "paultyng.asc"
  }
}

provider "hashicorp" "null" {
  registry {
    source = "hashicorp/null"
  }
}
```

Then run the tool using `tfstaticregistry` and it will fetch the specified providers and build a static site.

## TODO

* Support provider renaming, currently the source name and destination names need to match or files and paths get out of sync
* Manual provider source
* S3 server type
* Azure static site
* Google static site
