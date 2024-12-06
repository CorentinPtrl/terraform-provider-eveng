// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &labResource{}
	_ resource.ResourceWithConfigure = &labResource{}
)

// NewLabResource is a helper function to simplify the provider implementation.
func NewLabResource() resource.Resource {
	return &labResource{}
}

// labResource is the resource implementation.
type labResource struct {
	client *evengsdk.Client
}

// labResourceModel describes the resource data model.
type labResourceModel struct {
	FolderPath  basetypes.StringValue `tfsdk:"folder_path"`
	Path        basetypes.StringValue `tfsdk:"path"`
	Author      basetypes.StringValue `tfsdk:"author"`
	Body        basetypes.StringValue `tfsdk:"body"`
	Description basetypes.StringValue `tfsdk:"description"`
	Filename    basetypes.StringValue `tfsdk:"filename"`
	Name        string                `tfsdk:"name"`
	Version     basetypes.StringValue `tfsdk:"version"`
	Id          basetypes.StringValue `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *labResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lab"
}

// Configure sets the provider data for the resource.
func (r *labResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *labResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"folder_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path of the lab.",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "Path of the lab.",
			},
			"author": schema.StringAttribute{
				Optional:    true,
				Description: "Author of the lab.",
			},
			"body": schema.StringAttribute{
				Optional:    true,
				Description: "Body content of the lab.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the lab.",
			},
			"filename": schema.StringAttribute{
				Computed:    true,
				Description: "Filename of the lab.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the lab.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the lab in string format.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Id of the lab.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *labResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan labResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var path string
	if !plan.FolderPath.IsNull() {
		path = plan.FolderPath.ValueString()
	}
	path = path + "/" + plan.Name + ".unl"
	err := r.client.Lab.CreateLab(path, evengsdk.Lab{
		Author:      plan.Author.ValueString(),
		Body:        plan.Body.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create lab", err.Error())
		return
	}
	lab, err := r.client.Lab.GetLab(path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read lab", err.Error())
		return
	}
	plan.Path = basetypes.NewStringValue(path)
	plan.Filename = basetypes.NewStringValue(lab.Filename)
	plan.Version = basetypes.NewStringValue(lab.Version.String())
	plan.Id = basetypes.NewStringValue(lab.Id)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *labResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state labResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	lab, err := r.client.Lab.GetLab(state.Path.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Author = stringToBasetype(lab.Author)
	state.Body = stringToBasetype(lab.Body)
	state.Description = stringToBasetype(lab.Description)
	state.Filename = basetypes.NewStringValue(lab.Filename)
	state.Version = basetypes.NewStringValue(lab.Version.String())

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *labResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan labResourceModel
	var state labResourceModel
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

	err := r.MoveLab(&plan, &state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to move lab", err.Error())
		return
	}
	err = r.client.Lab.UpdateLab(state.Path.ValueString(), evengsdk.Lab{
		Name:        plan.Name,
		Author:      plan.Author.ValueString(),
		Body:        plan.Body.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update lab", err.Error())
		return
	}
	state.Path = basetypes.NewStringValue(plan.FolderPath.ValueString() + "/" + plan.Name + ".unl")
	lab, err := r.client.Lab.GetLab(state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read lab", err.Error())
		return
	}
	state.Author = stringToBasetype(lab.Author)
	state.Body = stringToBasetype(lab.Body)
	state.Description = stringToBasetype(lab.Description)
	state.Filename = basetypes.NewStringValue(lab.Filename)
	state.Version = basetypes.NewStringValue(lab.Version.String())
	state.Name = lab.Name
	state.FolderPath = plan.FolderPath

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *labResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state labResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Lab.DeleteLab(state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete lab", err.Error())
		return
	}
}

func (r *labResource) MoveLab(plan *labResourceModel, state *labResourceModel) error {
	if plan.FolderPath.ValueString() != state.FolderPath.ValueString() {
		path := plan.FolderPath.ValueString() + "/" + state.Name + ".unl"
		otherLab, err := r.client.Lab.GetLab(plan.FolderPath.ValueString() + "/" + state.Name + ".unl")
		if err == nil && otherLab.Id != state.Id.ValueString() {
			return fmt.Errorf("Lab already exists in the new folder")
		} else if err == nil && otherLab.Id == state.Id.ValueString() {
			state.Path = basetypes.NewStringValue(path)
			return nil
		}
		err = r.client.Lab.MoveLab(state.Path.ValueString(), plan.FolderPath.ValueString())
		if err != nil {
			return err
		}
		state.Path = basetypes.NewStringValue(path)
	}
	return nil
}

func stringToBasetype(s string) basetypes.StringValue {
	if s == "" {
		return basetypes.NewStringNull()
	}
	return basetypes.NewStringValue(s)
}
