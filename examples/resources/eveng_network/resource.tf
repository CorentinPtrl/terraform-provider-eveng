terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}

resource "eveng_lab" "example" {
  name   = "NetworkExample"
  author = "Corentin"
  body   = "Example of a lab with a network"
}

resource "eveng_network" "bridged" {
  lab_path = eveng_lab.example.path
  top      = 0
  left     = 0
  name     = "example_network"
  icon     = "01-Cloud-Default.svg"
  type     = "bridge"
}
