terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}

resource "eveng_folder" "example" {
  path = "/test"
}

resource "eveng_lab" "example" {
  folder_path = eveng_folder.example.path
  name        = "LabExampleWithFolder"
}
