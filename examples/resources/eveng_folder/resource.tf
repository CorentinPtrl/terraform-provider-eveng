terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}

resource "eveng_folder" "example" {
  path = "/example"
}
