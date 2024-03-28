package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martin-magakian/terraform-provider-beeswax/internal/beeswax-client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &userResource{}
	_ resource.ResourceWithConfigure = &userResource{}
)

// userResource is the resource implementation.
type userResource struct {
	client *beeswax.Client
}

// userResourceModel is the data the resource manipulates.
type userResourceModel struct {
	ID               types.Int64   `tfsdk:"id"`
	Email            types.String  `tfsdk:"email"`
	FirstName        types.String  `tfsdk:"first_name"`
	LastName         types.String  `tfsdk:"last_name"`
	RoleID           types.Int64   `tfsdk:"role_id"`
	AccountGroupIDs  []types.Int64 `tfsdk:"account_group_ids"`
	AccountID        types.Int64   `tfsdk:"account_id"`
	Active           types.Bool    `tfsdk:"active"`
	SuperUser        types.Bool    `tfsdk:"super_user"`
	AllAccountAccess types.Bool    `tfsdk:"all_account_access"`
}

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = defaultConfiguration(req.ProviderData, resp.Diagnostics)
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                 schema.Int64Attribute{Computed: true},
			"email":              schema.StringAttribute{Required: true},
			"first_name":         schema.StringAttribute{Required: true},
			"last_name":          schema.StringAttribute{Required: true},
			"role_id":            schema.Int64Attribute{Required: true},
			"account_group_ids":  schema.ListAttribute{Required: true, ElementType: types.Int64Type},
			"account_id":         schema.Int64Attribute{Optional: true},
			"active":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"super_user":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"all_account_access": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new user
	user := convertToUser(plan)
	userId, err := r.client.CreateUser(user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int64Value(userId)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get user from Beeswax API
	user, err := r.client.GetUser(state.ID.ValueInt64())
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

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	var state userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	diags2 := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update user
	user := convertToUser(plan)
	user.ID = state.ID.ValueInt64()
	err := r.client.UpdateUser(user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating user",
			"Could not update user, unexpected error: "+err.Error(),
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
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var plan userResourceModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete user
	err := r.client.DeleteUser(plan.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting user",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}
}

func convertToUser(plan userResourceModel) beeswax.User {
	return beeswax.User{
		ID:               plan.ID.ValueInt64(),
		SuperUser:        plan.SuperUser.ValueBool(),
		Email:            plan.Email.ValueString(),
		FirstName:        plan.FirstName.ValueString(),
		LastName:         plan.LastName.ValueString(),
		RoleID:           plan.RoleID.ValueInt64(),
		AccountID:        plan.AccountID.ValueInt64(),
		Active:           plan.Active.ValueBool(),
		AllAccountAccess: plan.AllAccountAccess.ValueBool(),
		AccountGroupIDs:  convertListInt(plan.AccountGroupIDs),
	}
}

func fillStateFromUser(state *userResourceModel, user beeswax.User) {
	state.ID = types.Int64Value(user.ID)
	state.SuperUser = types.BoolValue(user.SuperUser)
	state.Email = types.StringValue(user.Email)
	state.FirstName = types.StringValue(user.FirstName)
	state.LastName = types.StringValue(user.LastName)
	state.RoleID = types.Int64Value(user.RoleID)
	state.AccountID = types.Int64Value(user.AccountID)
	state.Active = types.BoolValue(user.Active)
	state.AllAccountAccess = types.BoolValue(user.AllAccountAccess)
	g := []types.Int64{}
	for _, id := range user.AccountGroupIDs {
		g = append(g, types.Int64Value(id))
	}
	state.AccountGroupIDs = g
}
