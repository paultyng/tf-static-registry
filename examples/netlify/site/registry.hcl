provider "paultyng" "unifi" {
  github {
    repository = "paultyng/terraform-provider-unifi
  }
}

provider "hashicorp" "null" {
  registry {
    source = "hashicorp/null"
  }
}