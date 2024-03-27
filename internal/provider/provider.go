package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martin-magakian/terraform-provider-beeswax/internal/beeswax-client"
)

// Ensure the provider satisfies the expected interfaces.
var (
	_ provider.Provider = &beeswaxProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &beeswaxProvider{
			version: version,
		}
	}
}

// beeswaxProvider is the provider implementation.
type beeswaxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// beeswaxProviderModel maps provider schema data to a Go type.
type beeswaxProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`
}

// Metadata returns the provider type name.
func (p *beeswaxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "beeswax"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *beeswaxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for Beeswax API. May also be provided via BEESWAX_HOST environment variable.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email to login to Beeswax API. May also be provided via BEESWAX_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password to login to Beeswax API. May also be provided via BEESWAX_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a beeswax API client for data sources and resources.
func (p *beeswaxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config beeswaxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check of unknown configuration values
	addUnknownDiagnostic := func(attr string) {
		resp.Diagnostics.AddAttributeError(
			path.Root(attr),
			"Unknown Beeswax API "+attr,
			"The provider cannot create the Beeswax API client as there is an unknown configuration value for the Beeswax API '"+attr+"'. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BEESWAX_"+strings.ToUpper(attr)+" environment variable.")
	}
	if config.Host.IsUnknown() {
		addUnknownDiagnostic("host")
	}
	if config.Email.IsUnknown() {
		addUnknownDiagnostic("email")
	}
	if config.Password.IsUnknown() {
		addUnknownDiagnostic("password")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Check of empty configuration values
	// Use environment variables as fallbacks
	host := os.Getenv("BEESWAX_HOST")
	email := os.Getenv("BEESWAX_EMAIL")
	password := os.Getenv("BEESWAX_PASSWORD")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if !config.Email.IsNull() {
		email = config.Email.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}
	addMissingDiagnostic := func(attr string) {
		resp.Diagnostics.AddAttributeError(
			path.Root(attr),
			"Missing Beeswax API "+attr,
			"The provider cannot create the Beeswax API client as there is a missing or empty value for the Beeswax API "+attr+". "+
				"Set the '"+attr+"' value in the configuration or use the BEESWAX_"+strings.ToUpper(attr)+" environment variable. "+
				"If either is already set, ensure the value is not empty.")
	}
	if host == "" {
		addMissingDiagnostic("host")
	}
	if email == "" {
		addMissingDiagnostic("email")
	}
	if password == "" {
		addMissingDiagnostic("password")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a Beeswax Client
	beeswaxClient := beeswax.NewClient(host, email, password)
	err := beeswaxClient.Login()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Beeswax API Client",
			"An unexpected error occurred when creating the Beeswax API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Beeswax Client Error: "+err.Error(),
		)
		return
	}

	// Make the Beeswax client available during DataSource and Resource type Configure methods.
	resp.DataSourceData = beeswaxClient
	resp.ResourceData = beeswaxClient
}

// DataSources defines the data sources implemented in the provider.
func (p *beeswaxProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewRoleDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *beeswaxProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewRoleResource,
	}
}
