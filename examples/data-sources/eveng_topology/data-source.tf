terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}

resource "eveng_lab" "example" {
  name   = "NodeLink"
  author = "Corentin"
}

resource "eveng_node" "node" {
  lab_path = eveng_lab.example.path
  name     = "switch_test_one"
  top      = 50
  left     = 50
  template = "viosl2"
  config   = "hostname switch_test"
  type     = "qemu"
}

resource "eveng_node" "test" {
  lab_path = eveng_lab.example.path
  name     = "switch_test_two"
  top      = 50
  left     = 500
  template = "viosl2"
  type     = "qemu"
}

resource "eveng_network" "bridged" {
  lab_path = eveng_lab.example.path
  top      = 100
  left     = 100
  name     = "test_network"
  icon     = "01-Cloud-Default.svg"
  type     = "bridge"
}

resource "eveng_node_link" "node" {
  lab_path       = eveng_lab.example.path
  source_node_id = eveng_node.node.id
  source_port    = "Gi0/1"
  target_node_id = eveng_node.test.id
  target_port    = "Gi0/1"
}

data "eveng_topology" "example" {
  lab_path = eveng_lab.example.path
}

output "topology" {
  value = data.eveng_topology.example.nodes
}
