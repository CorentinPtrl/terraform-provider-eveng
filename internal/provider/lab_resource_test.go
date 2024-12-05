// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLabResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLabResourceConfig("acceptance-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_lab.test", "path", "/acceptance-test.unl"),
					resource.TestCheckResourceAttr("eveng_lab.test", "name", "acceptance-test"),
					resource.TestCheckResourceAttr("eveng_lab.test", "author", "terraform-acctest"),
					resource.TestCheckResourceAttr("eveng_lab.test", "body", "terraform acceptance test"),
					resource.TestCheckResourceAttr("eveng_lab.test", "description", "terraform acceptance test"),
				),
			},
			// Update and Read testing
			{
				Config: testAccLabResourceConfig("acceptance-test-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_lab.test", "path", "/acceptance-test-update.unl"),
					resource.TestCheckResourceAttr("eveng_lab.test", "name", "acceptance-test-update"),
					resource.TestCheckResourceAttr("eveng_lab.test", "author", "terraform-acctest"),
					resource.TestCheckResourceAttr("eveng_lab.test", "body", "terraform acceptance test"),
					resource.TestCheckResourceAttr("eveng_lab.test", "description", "terraform acceptance test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "eveng_lab" "test" {
	name = %[1]q
	author = "terraform-acctest"
	body = "terraform acceptance test"
	description = "terraform acceptance test"
}
`, configurableAttribute)
}
