package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

type userDataSource struct {
	// share implementation with the resource
	userResource
}

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	r.client = defaultConfiguration(req.ProviderData, resp.Diagnostics)
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                 schema.Int64Attribute{Required: true},
			"email":              schema.StringAttribute{Computed: true}, // TODO: allow getting resource from email
			"first_name":         schema.StringAttribute{Computed: true},
			"last_name":          schema.StringAttribute{Computed: true},
			"role_id":            schema.Int64Attribute{Computed: true},
			"account_group_ids":  schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
			"account_id":         schema.Int64Attribute{Computed: true},
			"active":             schema.BoolAttribute{Computed: true},
			"super_user":         schema.BoolAttribute{Computed: true},
			"all_account_access": schema.BoolAttribute{Computed: true},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var state userResourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get user from Beeswax API
	user, err := d.client.GetUser(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Beeswax User",
			fmt.Sprintf("Could not read Beeswax User ID %d: %s", state.ID.ValueInt64(), err.Error()),
		)
		return
	}

	// Overwrite items with refreshed state
	fillStateFromUser(&state, user)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
