// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ datasource.DataSource              = &topologyDataSource{}
	_ datasource.DataSourceWithConfigure = &topologyDataSource{}
)

func NewTopologyDataSource() datasource.DataSource {
	return &topologyDataSource{}
}

type topologyDataSource struct {
	client *evengsdk.Client
}

type TopologyDataSourceModel struct {
	LabPath string        `tfsdk:"lab_path"`
	Nodes   types.Dynamic `tfsdk:"nodes"`
}

func (d *topologyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topology"
}

func (d *topologyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *topologyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"lab_path": schema.StringAttribute{
				Required:    true,
				Description: "Path of the lab.",
			},
			"nodes": schema.DynamicAttribute{
				Computed:    true,
				Description: "An array of nodes in the topology.",
			},
		},
	}
}

func (d *topologyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TopologyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	topology, err := d.client.Lab.GetTopology(state.LabPath)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var list []attr.Value
	var attributeTypes map[string]attr.Type
	topology = harmonizeMaps(topology)
	for _, node := range topology {
		var terraformType attr.Value
		terraformType, attributeTypes, err = createAttrValueFromMap(node)
		if err != nil {
			resp.Diagnostics.AddError("Failed to create dynamic value", err.Error())
			return
		}
		list = append(list, terraformType)
	}
	state.Nodes = basetypes.NewDynamicValue(basetypes.NewListValueMust(basetypes.ObjectType{AttrTypes: attributeTypes}, list))

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func createAttrValueFromMap(data map[string]interface{}) (attr.Value, map[string]attr.Type, error) {
	attributeTypes := map[string]attr.Type{}
	attributeValues := map[string]attr.Value{}

	for key, value := range data {
		attributeTypes[key] = basetypes.StringType{}
		attributeValues[key] = types.StringValue(fmt.Sprintf("%v", value))
	}

	objectValue, diag := types.ObjectValue(attributeTypes, attributeValues)
	if diag.HasError() {
		return objectValue, attributeTypes, fmt.Errorf("error creating object value: %v", diag.Errors())
	}

	return objectValue, attributeTypes, nil
}

func harmonizeMaps(maps []map[string]interface{}) []map[string]interface{} {
	if len(maps) == 0 {
		return maps
	}

	for i, m := range maps {
		for d, m2 := range maps {
			if i == d {
				continue
			}
			for k, _ := range m2 {
				if _, ok := m[k]; !ok {
					m[k] = ""
				}
			}
		}
	}
	return maps
}
