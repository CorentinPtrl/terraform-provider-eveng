// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/CorentinPtrl/evengsdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure EvengProvider satisfies various provider interfaces.
var _ provider.Provider = &EvengProvider{}
var _ provider.ProviderWithFunctions = &EvengProvider{}

// EvengProvider defines the provider implementation.
type EvengProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// EvengProviderModel describes the provider data model.
type EvengProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *EvengProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "eveng"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *EvengProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "The host of the Eveng API. (Can also be set with the EVE_HOST environment variable)",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "The username for the Eveng API. (Can also be set with the EVE_USER environment variable)",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The password for the Eveng API. (Can also be set with the EVE_PASSWORD environment variable)",
			},
		},
	}
}

func (p *EvengProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config EvengProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Eveng API Host",
			"The provider cannot create the Eveng API client as there is an unknown configuration value for the Eveng API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the EVE_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Eveng API Username",
			"The provider cannot create the Eveng API client as there is an unknown configuration value for the Eveng API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the EVE_USER environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Eveng API Password",
			"The provider cannot create the Eveng API client as there is an unknown configuration value for the Eveng API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the EVE_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("EVE_HOST")
	username := os.Getenv("EVE_USER")
	password := os.Getenv("EVE_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Eveng API Host",
			"The provider cannot create the Eveng API client as there is a missing or empty value for the Eveng API host. "+
				"Set the host value in the configuration or use the EVE_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Eveng API Username",
			"The provider cannot create the Eveng API client as there is a missing or empty value for the Eveng API username. "+
				"Set the username value in the configuration or use the EVE_USER environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Eveng API Password",
			"The provider cannot create the Eveng API client as there is a missing or empty value for the Eveng API password. "+
				"Set the password value in the configuration or use the EVE_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := evengsdk.NewBasicAuthClient(username, password, "0", host)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Eveng API client",
			"An error occurred while creating the Eveng API client. Please check the configuration values and try again.\n\n"+
				"Eveng Client Error: "+err.Error(),
		)
		return
	}

	// Make the Eveng client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *EvengProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFolderResource,
		NewLabResource,
		NewNodeResource,
		NewNetworkResource,
		NewNodeLinkResource,
		NewStartNodesResource,
	}
}

func (p *EvengProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFolderDataSource,
	}
}

func (p *EvengProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &EvengProvider{
			version: version,
		}
	}
}
