// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFolderResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFolderResourceConfig("/unit-acc-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_folder.test", "path", "/unit-acc-test"),
				),
			},
			// Update and Read testing
			{
				Config: testAccFolderResourceConfig("/unit-acc-test-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("eveng_folder.test", "path", "/unit-acc-test-update"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFolderResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "eveng_folder" "test" {
  path = %[1]q
}
`, configurableAttribute)
}
