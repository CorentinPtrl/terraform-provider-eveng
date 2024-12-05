// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNodeResourceConfig("acceptance-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_node.test", "lab_path", "/terraform-acceptance-test-node.unl"),
					resource.TestCheckResourceAttr("eveng_node.test", "name", "acceptance-test"),
					resource.TestCheckResourceAttr("eveng_node.test", "icon", "PC-2D-Desktop-Generic-S.svg"),
					resource.TestCheckResourceAttr("eveng_node.test", "ram", "1024"),
					resource.TestCheckResourceAttr("eveng_node.test", "cpu", "1"),
					resource.TestCheckResourceAttr("eveng_node.test", "ethernet", "4"),
					resource.TestCheckResourceAttr("eveng_node.test", "top", "0"),
					resource.TestCheckResourceAttr("eveng_node.test", "left", "0"),
				),
			},
			// Update and Read testing
			{
				Config: testAccNodeResourceConfig("acceptance-test-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_node.test", "lab_path", "/terraform-acceptance-test-node.unl"),
					resource.TestCheckResourceAttr("eveng_node.test", "name", "acceptance-test-update"),
					resource.TestCheckResourceAttr("eveng_node.test", "icon", "PC-2D-Desktop-Generic-S.svg"),
					resource.TestCheckResourceAttr("eveng_node.test", "ram", "1024"),
					resource.TestCheckResourceAttr("eveng_node.test", "cpu", "1"),
					resource.TestCheckResourceAttr("eveng_node.test", "ethernet", "4"),
					resource.TestCheckResourceAttr("eveng_node.test", "top", "0"),
					resource.TestCheckResourceAttr("eveng_node.test", "left", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNodeResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "eveng_lab" "test" {
	name = "terraform-acceptance-test-node"
	author = "terraform-acctest"
	body = "terraform acceptance test"
	description = "terraform acceptance test"
}

resource "eveng_node" "test" {
  lab_path = eveng_lab.test.path
  name = %[1]q
  template = "vpcs"
  type = "qemu"
}
`, configurableAttribute)
}
