terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}

resource "eveng_lab" "example" {
  name = "LabExample"
}
