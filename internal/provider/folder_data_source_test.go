// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEveFolderDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccFolderDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.eveng_folder.test", "path", "/"),
					resource.TestCheckResourceAttrSet("data.eveng_folder.test", "folders.#"),
					resource.TestCheckResourceAttrSet("data.eveng_folder.test", "labs.#"),
				),
			},
		},
	})
}

const testAccFolderDataSourceConfig = `
data "eveng_folder" "test" {
  path = "/"
}
`
