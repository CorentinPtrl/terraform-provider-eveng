terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}

data "eveng_folder" "example" {
  path = "/"
}

output "folder" {
  value = data.eveng_folder.example
}
