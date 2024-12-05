// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &networkResource{}
	_ resource.ResourceWithConfigure = &networkResource{}
)

// NewNetworkResource is a helper function to simplify the provider implementation.
func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

// networkResource is the resource implementation.
type networkResource struct {
	client *evengsdk.Client
}

// NetworkResourceModel describes the resource data model.
type NetworkResourceModel struct {
	LabPath types.String `tfsdk:"lab_path"`
	Id      types.Int64  `tfsdk:"id"`
	Left    types.Int64  `tfsdk:"left"`
	Name    types.String `tfsdk:"name"`
	Top     types.Int64  `tfsdk:"top"`
	Type    types.String `tfsdk:"type"`
	Icon    types.String `tfsdk:"icon"`
}

// Metadata returns the resource type name.
func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

// Configure sets the provider data for the resource.
func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Schema defines the schema for the resource.
func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"lab_path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the lab file.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Unique identifier of the network.",
			},
			"left": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Left position of the network.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network.",
			},
			"top": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Top position of the network.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of the network.",
			},
			"icon": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Icon representing the network.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network := r.NewNode(plan)
	err := r.client.Network.CreateNetwork(plan.LabPath.ValueString(), &network)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create network", err.Error())
		return
	}
	rnet, err := r.NewModel(plan.LabPath.ValueString(), network.Id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read network", err.Error())
		return
	}
	diags = resp.State.Set(ctx, rnet)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rnet, err := r.NewModel(state.LabPath.ValueString(), int(state.Id.ValueInt64()))
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, rnet)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NetworkResourceModel
	var state NetworkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network := r.NewNode(plan)
	network.Id = int(state.Id.ValueInt64())
	err := r.client.Network.UpdateNetwork(plan.LabPath.ValueString(), &network)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update network", err.Error())
		return
	}

	rnet, err := r.NewModel(plan.LabPath.ValueString(), network.Id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read network", err.Error())
		return
	}

	diags = resp.State.Set(ctx, rnet)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Network.DeleteNetwork(state.LabPath.ValueString(), int(state.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete network", err.Error())
		return
	}
}

func (r *networkResource) NewNode(model NetworkResourceModel) evengsdk.Network {
	network := evengsdk.Network{}
	if !model.Id.IsUnknown() {
		network.Id = int(model.Id.ValueInt64())
	}
	if !model.Left.IsUnknown() {
		network.Left = int(model.Left.ValueInt64())
	}
	if !model.Name.IsUnknown() {
		network.Name = model.Name.ValueString()
	}
	if !model.Top.IsUnknown() {
		network.Top = int(model.Top.ValueInt64())
	}
	if !model.Type.IsUnknown() {
		network.Type = model.Type.ValueString()
	}
	if !model.Icon.IsUnknown() {
		network.Icon = model.Icon.ValueString()
	}
	network.Visibility = "1"
	return network
}

func (r *networkResource) NewModel(labPath string, netId int) (NetworkResourceModel, error) {
	model := NetworkResourceModel{}
	model.LabPath = types.StringValue(labPath)
	model.Id = types.Int64Value(int64(netId))
	net, err := r.client.Network.GetNetwork(labPath, netId)
	if err != nil {
		return model, err
	}
	model.Left = types.Int64Value(int64(net.Left))
	model.Name = types.StringValue(net.Name)
	model.Top = types.Int64Value(int64(net.Top))
	model.Type = types.StringValue(net.Type)
	model.Icon = types.StringValue(net.Icon)
	return model, nil
}
