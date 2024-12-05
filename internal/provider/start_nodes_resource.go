// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &startNodesResource{}
	_ resource.ResourceWithConfigure = &startNodesResource{}
)

// NewStartNodesResource is a helper function to simplify the provider implementation.
func NewStartNodesResource() resource.Resource {
	return &startNodesResource{}
}

// startNodesResource is the resource implementation.
type startNodesResource struct {
	client *evengsdk.Client
}

// startNodesResourceModel describes the resource data model.
type startNodesResourceModel struct {
	LabPath   basetypes.StringValue `tfsdk:"lab_path"`
	StartTime basetypes.Int64Value  `tfsdk:"start_time"`
}

// Metadata returns the resource type name.
func (r *startNodesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_start_nodes"
}

// Configure sets the provider data for the resource.
func (r *startNodesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *startNodesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"lab_path": schema.StringAttribute{
				Required:    true,
				Description: "Path of the lab.",
			},
			"start_time": schema.Int64Attribute{
				Computed:    true,
				Description: "Time when the nodes were started.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *startNodesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan startNodesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.Lab.GetLab(plan.LabPath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read lab", err.Error())
		return
	}
	startTime, err := r.StartLab(plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to start nodes", err.Error())
		return
	}
	plan.StartTime = basetypes.NewInt64Value(startTime)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *startNodesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state startNodesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.StartTime = basetypes.NewInt64Null()
	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *startNodesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan startNodesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.Lab.GetLab(plan.LabPath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read lab", err.Error())
		return
	}
	startTime, err := r.StartLab(plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to start nodes", err.Error())
		return
	}
	plan.StartTime = basetypes.NewInt64Value(startTime)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *startNodesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *startNodesResource) StartLab(model startNodesResourceModel) (int64, error) {
	currentLab, err := r.client.GetAuth()
	if err != nil {
		return 0, err
	}
	if currentLab.Lab != model.LabPath.ValueString() && currentLab.Lab != "" {
		err = r.client.Node.StopNodes(currentLab.Lab)
		if err != nil {
			return 0, fmt.Errorf("Failed to stop nodes: %w", err)
		}
	}
	tries := 5
	delay := 15
	if currentLab.Lab == model.LabPath.ValueString() || currentLab.Lab == "" {
		tries = 0
	}
	for i := 0; i < tries; i++ {
		err = r.client.Lab.CloseLab()
		time.Sleep(time.Duration(delay) * time.Second)
	}
	if err != nil {
		return 0, fmt.Errorf("Failed to close lab: %w", err)
	}
	err = r.client.Node.StartNodes(model.LabPath.ValueString())
	if err != nil {
		return 0, fmt.Errorf("Failed to start nodes: %w", err)
	}
	return time.Now().Unix(), nil
}
