// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNetworkResourceConfig("acceptance-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_network.test", "lab_path", "/terraform-acceptance-test-network.unl"),
					resource.TestCheckResourceAttr("eveng_network.test", "name", "acceptance-test"),
					resource.TestCheckResourceAttr("eveng_network.test", "icon", "01-Cloud-Default.svg"),
					resource.TestCheckResourceAttr("eveng_network.test", "type", "bridge"),
					resource.TestCheckResourceAttr("eveng_network.test", "top", "0"),
					resource.TestCheckResourceAttr("eveng_network.test", "left", "0"),
				),
			},
			// Update and Read testing
			{
				Config: testAccNetworkResourceConfig("acceptance-test-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_network.test", "lab_path", "/terraform-acceptance-test-network.unl"),
					resource.TestCheckResourceAttr("eveng_network.test", "name", "acceptance-test-update"),
					resource.TestCheckResourceAttr("eveng_network.test", "icon", "01-Cloud-Default.svg"),
					resource.TestCheckResourceAttr("eveng_network.test", "type", "bridge"),
					resource.TestCheckResourceAttr("eveng_network.test", "top", "0"),
					resource.TestCheckResourceAttr("eveng_network.test", "left", "0")),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNetworkResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "eveng_lab" "test" {
	name = "terraform-acceptance-test-network"
	author = "terraform-acctest"
	body = "terraform acceptance test"
	description = "terraform acceptance test"
}

resource "eveng_network" "test" {
  lab_path = eveng_lab.test.path
  name = %[1]q
  icon = "01-Cloud-Default.svg"
  type = "bridge"
}
`, configurableAttribute)
}
