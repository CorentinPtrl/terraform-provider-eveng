terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}

resource "eveng_lab" "example" {
  name   = "NodeTest"
  author = "Corentin"
  body   = "Example of a lab with a node"
}

resource "eveng_node" "node" {
  lab_path = eveng_lab.example.path
  name     = "vpc"
  template = "vpcs"
  type     = "qemu"
}