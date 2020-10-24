terraform {
  required_providers {
    unifi = {
      source = "tf-static-registry-example.netlify.app/paultyng/unifi"
      version = "0.16.0"
    }
    new_null = {
      source = "tf-static-registry-example.netlify.app/hashicorp/null"
    }
    old_null = {
      source = "hashicorp/null"
    }
  }
}
