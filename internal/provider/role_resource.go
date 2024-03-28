package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martin-magakian/terraform-provider-beeswax/internal/beeswax-client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &roleResource{}
	_ resource.ResourceWithConfigure = &roleResource{}
)

// roleResource is the resource implementation.
type roleResource struct {
	client *beeswax.Client
}

// roleResourceModel is the data the resource manipulates.
type roleResourceModel struct {
	ID                   types.Int64               `tfsdk:"id"`
	Name                 types.String              `tfsdk:"name"`
	ParentRoleID         types.Int64               `tfsdk:"parent_role_id"`
	Archived             types.Bool                `tfsdk:"archived"`
	Notes                types.String              `tfsdk:"notes"`
	SharedAcrossAccounts types.Bool                `tfsdk:"shared_across_accounts"`
	Permissions          []permissionResourceModel `tfsdk:"permissions"`
	ReportIDs            []types.Int64             `tfsdk:"report_ids"`
}

type permissionResourceModel struct {
	ObjectType types.String `tfsdk:"object_type"`
	Permission types.Int64  `tfsdk:"permission"`
}

// NewroleResource is a helper function to simplify the provider implementation.
func NewRoleResource() resource.Resource {
	return &roleResource{}
}

// Metadata returns the resource type name.
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Configure adds the provider configured client to the resource.
func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = defaultConfiguration(req.ProviderData, resp.Diagnostics)
}

// Schema defines the schema for the resource.
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                     schema.Int64Attribute{Computed: true, Description: "Unique ID of the role"},
			"name":                   schema.StringAttribute{Required: true, Description: "Name of the role"}, // TODO: allow getting resource from name
			"parent_role_id":         schema.Int64Attribute{Required: true, Description: "The system role that determines which default permissions will be inherited"},
			"archived":               schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Archived roles cannot add new users"},
			"notes":                  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Free-form notes of up to 255 characters."},
			"shared_across_accounts": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "A role that can be shared across accounts, which can be enabled by all-accounts users."},
			"report_ids":             schema.ListAttribute{Optional: true, Computed: true, ElementType: types.Int64Type, Description: "List of IDs of reports users associated with this role should be able to access. A list of reports may be queried using /reporting/reports."},
			"permissions": schema.ListNestedAttribute{
				Required:    true,
				Description: "Object containing resource-level permissions for this Role",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"object_type": schema.StringAttribute{Required: true, Description: `The name of the resource, e.g. "advertiser" (note, these are singular)`},
						"permission":  schema.Int64Attribute{Required: true, Description: "4-bit integer determining Read (1), Create (2), Update (4) and Delete (8) rights for the resource. If a Permission is set to 1, the Role can only Read that type of object. If set to 3, the Role can Read and Create the object (1+2). When a Permission is set to 15 the Role has full rights to the object (1+2+4+8), if set to zero the Role has no rights."},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new role
	role := convertToRole(plan)
	roleId, err := r.client.CreateRole(role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role",
			"Could not create role, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int64Value(roleId)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get role from Beeswax API
	role, err := r.client.GetRole(state.ID.ValueInt64())
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

// Update updates the resource and sets the updated Terraform state on success.
func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan roleResourceModel
	var state roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	diags2 := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update role
	role := convertToRole(plan)
	role.ID = state.ID.ValueInt64()
	err := r.client.UpdateRole(role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating role",
			"Could not update role, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID // Keep the same ID

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var plan roleResourceModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete role
	err := r.client.DeleteRole(plan.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting role",
			"Could not delete role, unexpected error: "+err.Error(),
		)
		return
	}
}

func convertToRole(plan roleResourceModel) beeswax.Role {
	permissions := []beeswax.Permission{}
	for _, p := range plan.Permissions {
		permissions = append(permissions, beeswax.Permission{
			ObjectType: p.ObjectType.ValueString(),
			Permission: p.Permission.ValueInt64(),
		})
	}
	return beeswax.Role{
		ID:                   plan.ID.ValueInt64(),
		Name:                 plan.Name.ValueString(),
		ParentRoleID:         plan.ParentRoleID.ValueInt64(),
		Archived:             plan.Archived.ValueBool(),
		Notes:                plan.Notes.ValueString(),
		SharedAcrossAccounts: plan.SharedAcrossAccounts.ValueBool(),
		Permissions:          permissions,
		ReportIDs:            convertListInt(plan.ReportIDs),
	}
}

func fillStateFromRole(state *roleResourceModel, role beeswax.Role) {
	state.ID = types.Int64Value(role.ID)
	state.Name = types.StringValue(role.Name)
	state.ParentRoleID = types.Int64Value(role.ParentRoleID)
	state.Archived = types.BoolValue(role.Archived)
	state.Notes = types.StringValue(role.Notes)
	state.SharedAcrossAccounts = types.BoolValue(role.SharedAcrossAccounts)
	permissions := []permissionResourceModel{}

	for _, p := range role.Permissions {
		permissions = append(permissions, permissionResourceModel{
			ObjectType: types.StringValue(p.ObjectType),
			Permission: types.Int64Value(p.Permission),
		})
	}
	state.Permissions = permissions
}
