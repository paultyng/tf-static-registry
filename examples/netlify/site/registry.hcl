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
