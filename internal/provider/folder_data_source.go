// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &folderDataSource{}
	_ datasource.DataSourceWithConfigure = &folderDataSource{}
)

func NewFolderDataSource() datasource.DataSource {
	return &folderDataSource{}
}

type folderDataSource struct {
	client *evengsdk.Client
}

type FolderDataSourceModel struct {
	Path    string        `tfsdk:"path"`
	Folders []FolderModel `tfsdk:"folders"`
	Labs    []LabModel    `tfsdk:"labs"`
}

type FolderModel struct {
	Name string `tfsdk:"name"`
	Path string `tfsdk:"path"`
}

type LabModel struct {
	File   string `tfsdk:"file"`
	Path   string `tfsdk:"path"`
	Umtime int64  `tfsdk:"umtime"`
	Mtime  string `tfsdk:"mtime"`
}

func (d *folderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

func (d *folderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*evengsdk.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *evengsdk.Client, got %T. Report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *folderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Required: true,
			},
			"folders": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"path": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"labs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"file": schema.StringAttribute{
							Computed: true,
						},
						"path": schema.StringAttribute{
							Computed: true,
						},
						"umtime": schema.Int64Attribute{
							Computed: true,
						},
						"mtime": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *folderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state FolderDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	folders, err := d.client.Folder.GetFolder(state.Path)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	for _, folder := range folders.Folders {
		state.Folders = append(state.Folders, FolderModel{
			Name: folder.Name,
			Path: folder.Path,
		})
	}
	for _, lab := range folders.Labs {
		state.Labs = append(state.Labs, LabModel{
			File:   lab.File,
			Path:   lab.Path,
			Umtime: lab.Umtime,
			Mtime:  lab.Mtime,
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
