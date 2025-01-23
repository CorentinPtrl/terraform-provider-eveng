// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

type StyleResourceModel struct {
	Style           types.String  `tfsdk:"style"`
	Color           types.String  `tfsdk:"color"`
	SrcPos          types.Float32 `tfsdk:"srcpos"`
	DstPos          types.Float32 `tfsdk:"dstpos"`
	LinkStyle       types.String  `tfsdk:"linkstyle"`
	Width           types.Int32   `tfsdk:"width"`
	Label           types.String  `tfsdk:"label"`
	LabelPos        types.Float32 `tfsdk:"labelpos"`
	Stub            types.Int32   `tfsdk:"stub"`
	Curviness       types.Int32   `tfsdk:"curviness"`
	BezierCurviness types.Int32   `tfsdk:"beziercurviness"`
	Round           types.Int32   `tfsdk:"round"`
	Midpoint        types.Float32 `tfsdk:"midpoint"`
}

// NodeLinkResourceModel describes the resource data model.
type NodeLinkResourceModel struct {
	LabPath      types.String        `tfsdk:"lab_path"`
	NetworkId    types.Int64         `tfsdk:"network_id"`
	SourceNodeId types.Int64         `tfsdk:"source_node_id"`
	SourcePort   types.String        `tfsdk:"source_port"`
	TargetNodeId types.Int64         `tfsdk:"target_node_id"`
	TargetPort   types.String        `tfsdk:"target_port"`
	Style        *StyleResourceModel `tfsdk:"style"`
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
			"style": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Style of the link(Only for the Pro version of EVE-NG).",
				Attributes: map[string]schema.Attribute{
					"style": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("Solid"),
						Validators: []validator.String{
							stringvalidator.OneOf("Solid", "Dashed"),
						},
						Description: "Style of the link.",
					},
					"color": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("#3e7089"),
						Description: "Color of the link in hexadecimal format.",
					},
					"srcpos": schema.Float32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     float32default.StaticFloat32(0.15),
						Description: "Position of the source.",
					},
					"dstpos": schema.Float32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     float32default.StaticFloat32(0.85),
						Description: "Position of the destination.",
					},
					"linkstyle": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("Straight"),
						Validators: []validator.String{
							stringvalidator.OneOf("Straight", "Bezier", "Flowchart", "StateMachine"),
						},
						Description: "Style of the link.",
					},
					"width": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(2),
						Description: "Width of the link.",
					},
					"label": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						Description: "Label of the link.",
					},
					"labelpos": schema.Float32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     float32default.StaticFloat32(0.5),
						Description: "Position of the label.",
					},
					"stub": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(0),
						Description: "Stub of the link.",
					},
					"curviness": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(10),
						Description: "Curviness of the link.",
					},
					"beziercurviness": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(150),
						Description: "Bezier curviness of the link.",
					},
					"round": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int32default.StaticInt32(0),
						Description: "Roundness of the link.",
					},
					"midpoint": schema.Float32Attribute{
						Optional: true,
						Computed: true,
						Default:  float32default.StaticFloat32(0.5),
					},
				},
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

	if r.client.IsPro() {
		r.MakeNodeStyle(ctx, plan)
		rstyle := r.NewStyleModel(ctx, plan)
		plan.Style = &rstyle
	}
	state := NodeLinkResourceModel{
		LabPath:      plan.LabPath,
		NetworkId:    basetypes.NewInt64Value(id),
		SourceNodeId: plan.SourceNodeId,
		SourcePort:   plan.SourcePort,
		TargetNodeId: plan.TargetNodeId,
		TargetPort:   plan.TargetPort,
		Style:        plan.Style,
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

	if r.client.IsPro() && state.Style != nil {
		style := r.NewStyleModel(ctx, state)
		state.Style = &style
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
	if r.client.IsPro() {
		r.MakeNodeStyle(ctx, plan)
		rstyle := r.NewStyleModel(ctx, plan)
		plan.Style = &rstyle
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

func (r *nodeLinkResource) NewStyleModel(ctx context.Context, plan NodeLinkResourceModel) StyleResourceModel {
	return r.GetTopologyForTargetNode(ctx, plan)
}

func (r *nodeLinkResource) GetTopologyForTargetNode(ctx context.Context, plan NodeLinkResourceModel) StyleResourceModel {
	topology, err := r.client.Lab.GetTopology(plan.LabPath.ValueString())
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to get topology %s", err))
	}
	for _, node := range topology {
		node_source, ok := node["source"].(string)
		if !ok {
			continue
		}
		source_label, ok := node["source_label"].(string)
		if !ok {
			continue
		}
		if node_source == fmt.Sprintf("node%d", plan.TargetNodeId.ValueInt64()) && source_label == plan.TargetPort.ValueString() {
			style, ok := node["style"].(string)
			if style == "" || !ok {
				style = "Solid"
			}
			color, ok := node["color"].(string)
			if color == "" || !ok {
				color = "#3e7089"
			}
			srcpos, _ := strconv.ParseFloat(node["srcpos"].(string), 32)
			dstpos, _ := strconv.ParseFloat(node["dstpos"].(string), 32)
			linkstyle, ok := node["linkstyle"].(string)
			if linkstyle == "" || !ok {
				linkstyle = "Straight"
			}
			width, _ := strconv.Atoi(node["width"].(string))
			label, ok := node["label"].(string)
			if !ok {
				label = ""
			}
			labelpos, _ := strconv.ParseFloat(node["labelpos"].(string), 32)
			stub, _ := strconv.Atoi(node["stub"].(string))
			curviness, _ := strconv.Atoi(node["curviness"].(string))
			beziercurviness, _ := strconv.Atoi(node["beziercurviness"].(string))
			round, _ := strconv.Atoi(node["round"].(string))
			midpoint, _ := strconv.ParseFloat(node["midpoint"].(string), 32)
			return StyleResourceModel{
				Style:           basetypes.NewStringValue(style),
				Color:           basetypes.NewStringValue(color),
				SrcPos:          basetypes.NewFloat32Value(float32(srcpos)),
				DstPos:          basetypes.NewFloat32Value(float32(dstpos)),
				LinkStyle:       basetypes.NewStringValue(linkstyle),
				Width:           basetypes.NewInt32Value(int32(width)),
				Label:           basetypes.NewStringValue(label),
				LabelPos:        basetypes.NewFloat32Value(float32(labelpos)),
				Stub:            basetypes.NewInt32Value(int32(stub)),
				Curviness:       basetypes.NewInt32Value(int32(curviness)),
				BezierCurviness: basetypes.NewInt32Value(int32(beziercurviness)),
				Round:           basetypes.NewInt32Value(int32(round)),
				Midpoint:        basetypes.NewFloat32Value(float32(midpoint)),
			}
		}
	}
	return StyleResourceModel{}
}

func (r *nodeLinkResource) MakeNodeStyle(ctx context.Context, plan NodeLinkResourceModel) {
	if plan.Style == nil {
		return
	}
	style := evengsdk.Style{
		Style:           plan.Style.Style.ValueString(),
		Color:           plan.Style.Color.ValueString(),
		Srcpos:          plan.Style.SrcPos.ValueFloat32(),
		Dstpos:          plan.Style.DstPos.ValueFloat32(),
		Linkstyle:       plan.Style.LinkStyle.ValueString(),
		Width:           json.Number(strconv.Itoa(int(plan.Style.Width.ValueInt32()))),
		Label:           plan.Style.Label.ValueString(),
		Labelpos:        plan.Style.LabelPos.ValueFloat32(),
		Stub:            json.Number(strconv.Itoa(int(plan.Style.Stub.ValueInt32()))),
		Curviness:       json.Number(strconv.Itoa(int(plan.Style.Curviness.ValueInt32()))),
		Beziercurviness: json.Number(strconv.Itoa(int(plan.Style.BezierCurviness.ValueInt32()))),
		Round:           json.Number(strconv.Itoa(int(plan.Style.Round.ValueInt32()))),
		Midpoint:        plan.Style.Midpoint.ValueFloat32(),
	}
	err := r.client.Node.UpdateNodeInterfaceStyleByName(plan.LabPath.ValueString(), int(plan.TargetNodeId.ValueInt64()), plan.TargetPort.ValueString(), style)
	if err != nil {
		tflog.Error(context.Background(), fmt.Sprintf("Failed to update node interface style %s", err))
	}
}
