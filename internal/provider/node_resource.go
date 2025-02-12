// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &nodeResource{}
	_ resource.ResourceWithConfigure = &nodeResource{}
)

// NewNodeResource is a helper function to simplify the provider implementation.
func NewNodeResource() resource.Resource {
	return &nodeResource{}
}

// nodeResource is the resource implementation.
type nodeResource struct {
	client *evengsdk.Client
}

// nodeResourceModel describes the resource data model.
type nodeResourceModel struct {
	LabPath    types.String `tfsdk:"lab_path"`
	Console    types.String `tfsdk:"console"`
	Delay      types.Int64  `tfsdk:"delay"`
	Id         types.Int64  `tfsdk:"id"`
	Left       types.Int64  `tfsdk:"left"`
	Icon       types.String `tfsdk:"icon"`
	Image      types.String `tfsdk:"image"`
	Name       types.String `tfsdk:"name"`
	Ram        types.Int64  `tfsdk:"ram"`
	Template   types.String `tfsdk:"template"`
	Type       types.String `tfsdk:"type"`
	Top        types.Int64  `tfsdk:"top"`
	Url        types.String `tfsdk:"url"`
	Config     types.String `tfsdk:"config"`
	Cpu        types.Int64  `tfsdk:"cpu"`
	Ethernet   types.Int64  `tfsdk:"ethernet"`
	Interfaces types.Object `tfsdk:"interfaces"`
	Uuid       types.String `tfsdk:"uuid"`
}

type interfacesResourceModel struct {
	Serial   types.List `tfsdk:"serial"`
	Ethernet types.List `tfsdk:"ethernet"`
}

// Metadata returns the resource type name.
func (r *nodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

// Configure sets the provider data for the resource.
func (r *nodeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *nodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"lab_path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the lab file.",
			},
			"console": schema.StringAttribute{
				Computed:    true,
				Description: "Console type of the node.",
			},
			"delay": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Delay in milliseconds.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Unique Id of the node.",
			},
			"left": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Left position of the node.",
			},
			"icon": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Icon for the node.",
			},
			"image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Image associated with the node.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the node.",
			},
			"ram": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "RAM allocated to the node.",
			},
			"template": schema.StringAttribute{
				Required:    true,
				Description: "Template used for the node.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of the node.",
			},
			"top": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Top position of the node.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL associated with the node.",
			},
			"config": schema.StringAttribute{
				Optional:    true,
				Description: "Startup configuration of the node.",
			},
			"cpu": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of CPUs allocated to the node.",
			},
			"ethernet": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of Ethernet interfaces.",
			},
			"interfaces": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Interfaces of the node.",
				Attributes: map[string]schema.Attribute{
					"serial": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Serial interfaces.",
					},
					"ethernet": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Ethernet interfaces.",
					},
				},
			},
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the node.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *nodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	node, err := r.NewNode(plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create node", err.Error())
		return
	}
	err = r.client.Node.CreateNode(plan.LabPath.ValueString(), &node)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create node", err.Error())
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Created node %d", node.Id))
	_, err = r.client.Node.GetNodeConfig(plan.LabPath.ValueString(), node.Id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get node config", err.Error())
		return
	}
	err = r.client.Node.UpdateNodeConfig(plan.LabPath.ValueString(), node.Id, plan.Config.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update node config", err.Error())
		return
	}
	node.Config = "1"
	err = r.client.Node.UpdateNode(plan.LabPath.ValueString(), &node)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update node config", err.Error())
		return
	}
	ints, err := r.NewInterfaceModel(plan.LabPath.ValueString(), node.Id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get node interfaces", err.Error())
		return
	}
	state, err := r.NewNodeModel(plan.LabPath.ValueString(), node.Id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get node", err.Error())
		return
	}
	objectValue, diags := types.ObjectValueFrom(ctx, ints.AttributeTypes(), ints)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Interfaces = objectValue
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *nodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.NewNodeModel(state.LabPath.ValueString(), int(state.Id.ValueInt64()))
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	ints, err := r.NewInterfaceModel(state.LabPath.ValueString(), int(state.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get node interfaces", err.Error())
		return
	}
	objectValue, diags := types.ObjectValueFrom(ctx, ints.AttributeTypes(), ints)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Interfaces = objectValue
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *nodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan nodeResourceModel
	var state nodeResourceModel
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

	node, err := r.NewNode(plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create node", err.Error())
		return
	}
	node.Id = int(state.Id.ValueInt64())
	node.Config = "1"
	err = r.client.Node.UpdateNodeConfig(plan.LabPath.ValueString(), node.Id, plan.Config.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update node config", err.Error())
		return
	}
	err = r.client.Node.UpdateNode(plan.LabPath.ValueString(), &node)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update node", err.Error())
		return
	}
	state, err = r.NewNodeModel(plan.LabPath.ValueString(), int(state.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get node", err.Error())
		return
	}
	ints, err := r.NewInterfaceModel(state.LabPath.ValueString(), int(state.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get node interfaces", err.Error())
		return
	}
	objectValue, diags := types.ObjectValueFrom(ctx, ints.AttributeTypes(), ints)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Interfaces = objectValue
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *nodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Node.DeleteNode(state.LabPath.ValueString(), int(state.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete node", err.Error())
		return
	}
}

func (r *nodeResource) NewNode(model nodeResourceModel) (evengsdk.Node, error) {
	tmpl, err := r.client.Node.GetTemplate(model.Template.ValueString())
	if err != nil {
		return evengsdk.Node{}, err
	}
	node := evengsdk.Node{}
	if !model.Console.IsUnknown() {
		node.Console = model.Console.ValueString()
	}
	if !model.Delay.IsUnknown() {
		node.Delay = int(model.Delay.ValueInt64())
	}
	if !model.Left.IsUnknown() {
		node.Left = int(model.Left.ValueInt64())
	}
	if !model.Icon.IsUnknown() {
		node.Icon = model.Icon.ValueString()
	} else {
		if icon, ok := tmpl["options"].(map[string]interface{})["icon"].(map[string]interface{})["value"].(string); ok {
			node.Icon = icon
		}
	}
	if !model.Image.IsUnknown() {
		node.Image = model.Image.ValueString()
	}
	node.Name = model.Name.ValueString()
	if !model.Ram.IsUnknown() {
		node.Ram = int(model.Ram.ValueInt64())
	}
	if !model.Template.IsUnknown() {
		node.Template = model.Template.ValueString()
	}
	if !model.Type.IsUnknown() {
		node.Type = model.Type.ValueString()
	}
	if !model.Top.IsUnknown() {
		node.Top = int(model.Top.ValueInt64())
	}
	if !model.Cpu.IsUnknown() {
		node.Cpu = int(model.Cpu.ValueInt64())
	}
	if !model.Ethernet.IsUnknown() {
		node.Ethernet = int(model.Ethernet.ValueInt64())
	} else {
		// Check if the template has an ethernet option
		_, ok := tmpl["options"].(map[string]interface{})["ethernet"]

		if ok {
			if ethernet, ok := tmpl["options"].(map[string]interface{})["ethernet"].(map[string]interface{})["value"].(float64); ok {
				node.Ethernet = int(ethernet)
			}
		}
	}
	return node, nil
}

func (r *nodeResource) NewNodeModel(labPath string, nodeId int) (nodeResourceModel, error) {
	node, err := r.client.Node.GetNode(labPath, nodeId)
	if err != nil {
		return nodeResourceModel{}, err
	}
	model := nodeResourceModel{}
	model.Console = types.StringValue(node.Console)
	model.Delay = types.Int64Value(int64(node.Delay))
	model.Left = types.Int64Value(int64(node.Left))
	model.Icon = types.StringValue(node.Icon)
	model.Image = types.StringValue(node.Image)
	model.Name = types.StringValue(node.Name)
	model.Ram = types.Int64Value(int64(node.Ram))
	model.Template = types.StringValue(node.Template)
	model.Type = types.StringValue(node.Type)
	model.Top = types.Int64Value(int64(node.Top))
	model.Url = types.StringValue(node.Url)
	model.Cpu = types.Int64Value(int64(node.Cpu))
	model.Ethernet = types.Int64Value(int64(node.Ethernet))
	model.Uuid = types.StringValue(node.Uuid)
	model.Id = types.Int64Value(int64(node.Id))
	config, err := r.client.Node.GetNodeConfig(labPath, nodeId)
	if err != nil {
		return nodeResourceModel{}, err
	}
	if config != "" {
		model.Config = types.StringValue(config)
	}
	model.LabPath = types.StringValue(labPath)
	return model, nil
}

func (r *nodeResource) NewInterfaceModel(labPath string, nodeId int) (interfacesResourceModel, error) {
	interfaces, err := r.client.Node.GetNodeInterfaces(labPath, nodeId)
	if err != nil {
		return interfacesResourceModel{}, err
	}
	model := interfacesResourceModel{}
	var serialInts []attr.Value
	for _, s := range interfaces.Serial {
		serialInts = append(serialInts, types.StringValue(s.Name))
	}
	serial, diags := types.ListValue(types.StringType, serialInts)
	if diags.HasError() {
		return model, errors.New("Failed to create serial interfaces list")
	}
	var ethernetInts []attr.Value
	for _, e := range interfaces.Ethernet {
		ethernetInts = append(ethernetInts, types.StringValue(e.Name))
	}
	ethernet, diags := types.ListValue(types.StringType, ethernetInts)
	if diags.HasError() {
		return model, errors.New("Failed to create ethernet interfaces list")
	}
	model.Serial = serial
	model.Ethernet = ethernet
	return model, nil
}

func (m interfacesResourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"serial":   types.ListType{ElemType: types.StringType},
		"ethernet": types.ListType{ElemType: types.StringType},
	}
}
