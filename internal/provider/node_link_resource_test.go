// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodeLinkNetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNodeLinkNetResourceConfig("e0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_node_link.test", "lab_path", "/terraform-acceptance-test-node-link.unl"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "network_id", "1"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "source_node_id", "1"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "source_port", "e0"),
				),
			},
			// Update and Read testing
			{
				Config: testAccNodeLinkNetResourceConfig("e1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_node_link.test", "lab_path", "/terraform-acceptance-test-node-link.unl"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "network_id", "1"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "source_node_id", "1"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "source_port", "e1")),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccNodeLinkNodeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNodeLinkNodeResourceConfig("e0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_node_link.test", "lab_path", "/terraform-acceptance-test-node-link.unl"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "network_id", "1"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "source_port", "e0"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "target_port", "e0"),
				),
			},
			// Update and Read testing
			{
				Config: testAccNodeLinkNodeResourceConfig("e1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_node_link.test", "lab_path", "/terraform-acceptance-test-node-link.unl"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "network_id", "1"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "source_port", "e1"),
					resource.TestCheckResourceAttr("eveng_node_link.test", "target_port", "e1")),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNodeLinkNetResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "eveng_lab" "test" {
	name = "terraform-acceptance-test-node-link"
	author = "terraform-acctest"
	body = "terraform acceptance test"
	description = "terraform acceptance test"
}

resource "eveng_network" "test" {
  lab_path = eveng_lab.test.path
  name = "acceptance-test-node-link"
  icon = "01-Cloud-Default.svg"
  type = "bridge"
}

resource "eveng_node" "test" {
  lab_path = eveng_lab.test.path
  name = "acceptance-test-vpc"
  template = "vpcs"
  type = "qemu"
}

resource "eveng_node_link" "test" {
  lab_path = eveng_lab.test.path
  network_id = eveng_network.test.id
  source_node_id = eveng_node.test.id
  source_port = %[1]q
}

`, configurableAttribute)
}

func testAccNodeLinkNodeResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "eveng_lab" "test" {
	name = "terraform-acceptance-test-node-link"
	author = "terraform-acctest"
	body = "terraform acceptance test"
	description = "terraform acceptance test"
}

resource "eveng_node" "test" {
  count = 2
  lab_path = eveng_lab.test.path
  name = "acceptance-test-vpc"
  template = "vpcs"
  type = "qemu"
}

resource "eveng_node_link" "test" {
  lab_path = eveng_lab.test.path
  source_node_id = eveng_node.test[0].id
  source_port = %[1]q
  target_node_id = eveng_node.test[1].id
  target_port = %[1]q
}

`, configurableAttribute)
}
