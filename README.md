# terraform-provider-eveng

This repository is a Terraform provider for EVE-NG (Emulated Virtual Environment Next Generation). It allows you to manage EVE-NG resources using Terraform.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules). Please see the Go documentation for the most up-to-date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the Provider

To use the provider, add it to your Terraform configuration as follows:

```terraform
terraform {
  required_providers {
    eveng = {
      source = "CorentinPtrl/eveng"
    }
  }
}

provider "eveng" {}
```

## Example Usage

```terraform
resource "eveng_lab" "example" {
  name = "LabExample"
}

resource "eveng_network" "bridged" {
  lab_path = eveng_lab.example.path
  top      = 0
  left     = 0
  name     = "example_network"
  icon     = "01-Cloud-Default.svg"
  type     = "bridge"
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

resource "eveng_node_link" "node" {
  lab_path       = eveng_lab.example.path
  network_id     = eveng_network.bridged.id
  source_node_id = eveng_node.node.id
  source_port    = "Gi0/1"
}

resource "eveng_start_nodes" "start" {
  lab_path   = eveng_lab.example.path
  depends_on = [eveng_node_link.node]
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources.

```shell
make testacc
```