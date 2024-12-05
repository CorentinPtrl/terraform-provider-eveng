// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &nodeLinkResource{}
	_ resource.ResourceWithConfigure = &nodeLinkResource{}
)

// NewNodeLinkResource is a helper function to simplify the provider implementation.
func NewNodeLinkResource() resource.Resource {
	return &nodeLinkResource{}
}

// nodeLinkResource is the resource implementation.
type nodeLinkResource struct {
	client *evengsdk.Client
}

// NodeLinkResourceModel describes the resource data model.
type NodeLinkResourceModel struct {
	LabPath      types.String `tfsdk:"lab_path"`
	NetworkId    types.Int64  `tfsdk:"network_id"`
	SourceNodeId types.Int64  `tfsdk:"source_node_id"`
	SourcePort   types.String `tfsdk:"source_port"`
	TargetNodeId types.Int64  `tfsdk:"target_node_id"`
	TargetPort   types.String `tfsdk:"target_port"`
}

// Metadata returns the resource type name.
func (r *nodeLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_link"
}

// Configure sets the provider data for the resource.
func (r *nodeLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *nodeLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"lab_path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the lab file.",
			},
			"network_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("target_node_id"),
					}...),
				},
				Description: "ID of the network.",
			},
			"source_node_id": schema.Int64Attribute{
				Required:    true,
				Description: "ID of the source node.",
			},
			"source_port": schema.StringAttribute{
				Required:    true,
				Description: "Source port.",
			},
			"target_node_id": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("network_id"),
					}...),
					int64validator.AlsoRequires(path.Expressions{
						path.MatchRoot("target_port"),
					}...),
				},
				Description: "ID of the target node.",
			},
			"target_port": schema.StringAttribute{
				Optional:    true,
				Description: "Target port.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *nodeLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NodeLinkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.SourceNodeId.ValueInt64() == plan.TargetNodeId.ValueInt64() {
		resp.Diagnostics.AddError("Cannot link a node to itself", "source and target node IDs are the same")
		return
	}

	var id int64
	var err error
	if !plan.NetworkId.IsUnknown() {
		id, err = r.MakeNodeLinkNet(plan, NodeLinkResourceModel{})
	} else {
		id, err = r.MakeNodeLinkNode(plan, NodeLinkResourceModel{})
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create node link (isNet=%t)", !plan.NetworkId.IsNull()), err.Error())
		return
	}
	tflog.Info(ctx, "Created node link", map[string]interface{}{
		"lab_path":   plan.LabPath.ValueString(),
		"network_id": id,
	})

	if id == 0 {
		resp.Diagnostics.AddError("Failed to create node link", "network ID is 0")
		return
	}

	state := NodeLinkResourceModel{
		LabPath:      plan.LabPath,
		NetworkId:    basetypes.NewInt64Value(id),
		SourceNodeId: plan.SourceNodeId,
		SourcePort:   plan.SourcePort,
		TargetNodeId: plan.TargetNodeId,
		TargetPort:   plan.TargetPort,
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *nodeLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NodeLinkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var recreate bool
	var err error
	if state.TargetNodeId.IsNull() {
		state, err, recreate = r.NewNodeLinkModelNet(state)
	} else {
		state, err, recreate = r.NewNodeLinkModelNode(state)
	}
	if recreate {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Failed to read node link", err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *nodeLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NodeLinkResourceModel
	var state NodeLinkResourceModel
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

	if plan.SourceNodeId.ValueInt64() == plan.TargetNodeId.ValueInt64() {
		resp.Diagnostics.AddError("Cannot link a node to itself", "source and target node IDs are the same")
		return
	}

	if !plan.TargetNodeId.IsNull() && !state.NetworkId.IsNull() && state.TargetNodeId.IsNull() {
		tflog.Info(ctx, "Node Link Changed from Net to Node")
		state.NetworkId = basetypes.NewInt64Unknown()
	}

	var id int64
	var err error
	if !plan.NetworkId.IsUnknown() {
		id, err = r.MakeNodeLinkNet(plan, state)
	} else {
		id, err = r.MakeNodeLinkNode(plan, state)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to update node link", err.Error())
		return
	}
	plan.NetworkId = basetypes.NewInt64Value(id)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *nodeLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NodeLinkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if state.NetworkId.IsUnknown() {
		return
	}
	if !state.TargetNodeId.IsNull() {
		err := r.client.Network.DeleteNetwork(state.LabPath.ValueString(), int(state.NetworkId.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Failed to delete node link", err.Error())
			return
		}
	} else {
		err := r.ensureInterfaceDeleted(state.LabPath.ValueString(), int(state.SourceNodeId.ValueInt64()), state.SourcePort.ValueString(), int(state.NetworkId.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Failed to delete node link", err.Error())
			return
		}
	}
}

func (r *nodeLinkResource) MakeNodeLinkNet(plan NodeLinkResourceModel, state NodeLinkResourceModel) (int64, error) {
	if ((plan.SourceNodeId.ValueInt64() != state.SourceNodeId.ValueInt64()) || plan.SourcePort.ValueString() != state.SourcePort.ValueString()) && state.SourceNodeId.ValueInt64() != 0 {
		err := r.ensureInterfaceDeleted(plan.LabPath.ValueString(), int(state.SourceNodeId.ValueInt64()), state.SourcePort.ValueString(), int(state.NetworkId.ValueInt64()))
		if err != nil {
			return plan.NetworkId.ValueInt64(), err
		}
	}
	err := r.client.Node.UpdateNodeInterfaceName(plan.LabPath.ValueString(), int(plan.SourceNodeId.ValueInt64()), plan.SourcePort.ValueString(), int(plan.NetworkId.ValueInt64()))
	if err != nil {
		return plan.NetworkId.ValueInt64(), err
	}
	return plan.NetworkId.ValueInt64(), nil
}

func (r *nodeLinkResource) NewNodeLinkModelNet(state NodeLinkResourceModel) (NodeLinkResourceModel, error, bool) {
	model := state
	_, err := r.client.Network.GetNetwork(state.LabPath.ValueString(), int(state.NetworkId.ValueInt64()))
	if err != nil {
		return model, err, true
	}
	_, err = r.client.Node.GetNode(state.LabPath.ValueString(), int(state.SourceNodeId.ValueInt64()))
	if err != nil {
		model.SourceNodeId = basetypes.NewInt64Value(0)
		model.SourcePort = basetypes.NewStringValue("")
		return model, fmt.Errorf("source node %d not found", state.SourceNodeId.ValueInt64()), false
	}
	if state.SourcePort.ValueString() == "" {
		return model, nil, false
	}
	_, sourceInt, err := r.client.Node.GetNodeInterface(state.LabPath.ValueString(), int(state.SourceNodeId.ValueInt64()), state.SourcePort.ValueString())
	if err != nil {
		model.SourcePort = basetypes.NewStringValue("")
		return model, err, false
	}
	if sourceInt.NetworkId != int(state.NetworkId.ValueInt64()) {
		model.SourcePort = basetypes.NewStringValue("")
		return model, nil, false
	}
	return model, nil, false
}

func (r *nodeLinkResource) MakeNodeLinkNode(plan NodeLinkResourceModel, state NodeLinkResourceModel) (int64, error) {
	if ((plan.SourceNodeId.ValueInt64() != state.SourceNodeId.ValueInt64()) || plan.SourcePort.ValueString() != state.SourcePort.ValueString()) && state.SourceNodeId.ValueInt64() != 0 {
		err := r.ensureInterfaceDeleted(plan.LabPath.ValueString(), int(state.SourceNodeId.ValueInt64()), state.SourcePort.ValueString(), int(state.NetworkId.ValueInt64()))
		if err != nil {
			return state.NetworkId.ValueInt64(), err
		}
	}
	if ((plan.TargetNodeId.ValueInt64() != state.TargetNodeId.ValueInt64()) || plan.TargetPort.ValueString() != state.TargetPort.ValueString()) && state.TargetNodeId.ValueInt64() != 0 {
		err := r.ensureInterfaceDeleted(plan.LabPath.ValueString(), int(state.TargetNodeId.ValueInt64()), state.TargetPort.ValueString(), int(state.NetworkId.ValueInt64()))
		if err != nil {
			return state.NetworkId.ValueInt64(), err
		}
	}
	sourceIndex, _, err := r.client.Node.GetNodeInterface(plan.LabPath.ValueString(), int(plan.SourceNodeId.ValueInt64()), plan.SourcePort.ValueString())
	if err != nil {
		return state.NetworkId.ValueInt64(), err
	}
	targetIndex, _, err := r.client.Node.GetNodeInterface(plan.LabPath.ValueString(), int(plan.TargetNodeId.ValueInt64()), plan.TargetPort.ValueString())
	if err != nil {
		return state.NetworkId.ValueInt64(), err
	}
	network, err := r.createOrUpdateNetwork(plan.LabPath.ValueString(), int(state.NetworkId.ValueInt64()), strconv.Itoa(int(plan.SourceNodeId.ValueInt64()))+"_"+strconv.Itoa(sourceIndex)+"_"+strconv.Itoa(int(plan.TargetNodeId.ValueInt64()))+"_"+strconv.Itoa(targetIndex))
	if err != nil {
		return int64(network.Id), err
	}
	err = r.client.Node.UpdateNodeInterfaceName(plan.LabPath.ValueString(), int(plan.SourceNodeId.ValueInt64()), plan.SourcePort.ValueString(), network.Id)
	if err != nil {
		return int64(network.Id), err
	}
	err = r.client.Node.UpdateNodeInterfaceName(plan.LabPath.ValueString(), int(plan.TargetNodeId.ValueInt64()), plan.TargetPort.ValueString(), network.Id)
	if err != nil {
		return int64(network.Id), err
	}
	network.Visibility = "0"
	err = r.client.Network.UpdateNetwork(plan.LabPath.ValueString(), &network)
	return int64(network.Id), err
}

func (r *nodeLinkResource) NewNodeLinkModelNode(state NodeLinkResourceModel) (NodeLinkResourceModel, error, bool) {
	model := state
	_, err := r.client.Network.GetNetwork(state.LabPath.ValueString(), int(state.NetworkId.ValueInt64()))
	if err != nil {
		return model, err, true
	}
	_, err = r.client.Node.GetNode(state.LabPath.ValueString(), int(state.SourceNodeId.ValueInt64()))
	if err != nil {
		model.SourceNodeId = basetypes.NewInt64Value(0)
		model.SourcePort = basetypes.NewStringValue("")
		return model, err, false
	}
	_, err = r.client.Node.GetNode(state.LabPath.ValueString(), int(state.TargetNodeId.ValueInt64()))
	if err != nil {
		model.TargetNodeId = basetypes.NewInt64Value(0)
		model.TargetPort = basetypes.NewStringValue("")
		return model, err, false
	}
	_, sourceInt, err := r.client.Node.GetNodeInterface(state.LabPath.ValueString(), int(state.SourceNodeId.ValueInt64()), state.SourcePort.ValueString())
	if err != nil {
		model.SourcePort = basetypes.NewStringValue("")
		return model, err, false
	}
	if sourceInt.NetworkId != int(state.NetworkId.ValueInt64()) {
		model.SourcePort = basetypes.NewStringValue("")
		return model, fmt.Errorf("source port %s is not connected to network %d", state.SourcePort.ValueString(), state.NetworkId.ValueInt64()), false
	}
	_, targetInt, err := r.client.Node.GetNodeInterface(state.LabPath.ValueString(), int(state.TargetNodeId.ValueInt64()), state.TargetPort.ValueString())
	if err != nil {
		model.SourcePort = basetypes.NewStringValue("")
		return model, err, false
	}
	if targetInt.NetworkId != int(state.NetworkId.ValueInt64()) {
		model.TargetPort = basetypes.NewStringValue("")
		return model, fmt.Errorf("target port %s is not connected to network %d", state.TargetPort.ValueString(), state.NetworkId.ValueInt64()), false
	}
	return model, nil, false
}

func (r *nodeLinkResource) ensureInterfaceDeleted(labPath string, nodeId int, port string, networkId int) error {
	_, inter, err := r.client.Node.GetNodeInterface(labPath, nodeId, port)
	if err != nil {
		return err
	}
	if inter.NetworkId == networkId {
		err = r.client.Node.UpdateNodeInterfaceName(labPath, nodeId, port, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *nodeLinkResource) createOrUpdateNetwork(labPath string, networkId int, netName string) (evengsdk.Network, error) {
	_, err := r.client.Network.GetNetwork(labPath, networkId)
	network := &evengsdk.Network{
		Id:         networkId,
		Left:       0,
		Top:        0,
		Name:       netName,
		Type:       "bridge",
		Visibility: "1",
		Icon:       "lan.png",
	}
	if err != nil {
		network.Id = 0
		err = r.client.Network.CreateNetwork(labPath, network)
		return *network, err
	} else {
		err = r.client.Network.UpdateNetwork(labPath, network)
		return *network, err
	}
}
