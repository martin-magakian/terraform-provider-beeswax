package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martin-magakian/terraform-provider-beeswax/internal/beeswax-client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &rolesDataSource{}
	_ datasource.DataSourceWithConfigure = &rolesDataSource{}
)

type rolesDataSource struct {
	client *beeswax.Client
}

func NewRolesDataSource() datasource.DataSource {
	return &rolesDataSource{}
}

// roleResourceModel is the data the resource manipulates.
type rolesResourceModel struct {
	Roles []liteRoleResourceModel `tfsdk:"roles"`
}

type liteRoleResourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (r *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	r.client = defaultConfiguration(req.ProviderData, resp.Diagnostics)
}

func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"roles": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Role available on Beeswax API",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.Int64Attribute{Computed: true, Description: "Unique ID of the role"},
						"name": schema.StringAttribute{Computed: true, Description: "Name of the role"},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var state rolesResourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get role from Beeswax API
	roles, err := d.client.GetRoles()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Beeswax role",
			fmt.Sprintf("Could not read Beeswax roles: %s", err.Error()),
		)
		return
	}

	// Overwrite items with refreshed state
	for _, role := range roles {
		state.Roles = append(state.Roles, liteRoleResourceModel{
			ID:   types.Int64Value(role.ID),
			Name: types.StringValue(role.Name),
		})
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
