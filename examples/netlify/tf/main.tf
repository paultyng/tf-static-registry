terraform {
  required_providers {
    unifi = {
      source = "tf-static-registry-example.netlify.app/paultyng/unifi"
      version = "0.16.0"
    }
    new-null = {
      source = "tf-static-registry-example.netlify.app/hashicorp/null"
    }
    old-null = {
      source = "hashicorp/null"
    }
  }
}
