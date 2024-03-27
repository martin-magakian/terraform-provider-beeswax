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
	_ datasource.DataSource              = &roleDataSource{}
	_ datasource.DataSourceWithConfigure = &roleDataSource{}
)

type roleDataSource struct {
	// share implementation with the resource
	roleResource
}

func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	r.client = defaultConfiguration(req.ProviderData, resp.Diagnostics)
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                     schema.Int64Attribute{Required: true, Description: "Unique ID of the role"},
			"name":                   schema.StringAttribute{Computed: true, Description: "Name of the role"}, // TODO: allow getting resource from name
			"parent_role_id":         schema.Int64Attribute{Computed: true, Description: "The system role that determines which default permissions will be inherited"},
			"archived":               schema.BoolAttribute{Computed: true, Description: "Archived roles cannot add new users"},
			"notes":                  schema.StringAttribute{Computed: true, Description: "Free-form notes of up to 255 characters."},
			"shared_across_accounts": schema.BoolAttribute{Computed: true, Description: "A role that can be shared across accounts, which can be enabled by all-accounts users."},
			"report_ids":             schema.ListAttribute{Computed: true, ElementType: types.Int64Type, Description: "List of IDs of reports users associated with this role should be able to access. A list of reports may be queried using /reporting/reports."},
			"permissions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Object containing resource-level permissions for this Role",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"object_type": schema.StringAttribute{Computed: true, Description: `The name of the resource, e.g. "advertiser" (note, these are singular)`},
						"permission":  schema.Int64Attribute{Computed: true, Description: "4-bit integer determining Read (1), Create (2), Update (4) and Delete (8) rights for the resource. If a Permission is set to 1, the Role can only Read that type of object. If set to 3, the Role can Read and Create the object (1+2). When a Permission is set to 15 the Role has full rights to the object (1+2+4+8), if set to zero the Role has no rights."},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var state roleResourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get role from Beeswax API
	role, err := d.client.GetRole(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Beeswax role",
			fmt.Sprintf("Could not read Beeswax role ID %d: %s", state.ID.ValueInt64(), err.Error()),
		)
		return
	}

	// Overwrite items with refreshed state
	fillStateFromRole(&state, role)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
