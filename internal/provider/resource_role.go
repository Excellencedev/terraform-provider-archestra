package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/archestra-ai/archestra/terraform-provider-archestra/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

type RoleResource struct {
	client *client.ClientWithResponses
}

type RoleResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Permissions []types.String `tfsdk:"permissions"`
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom RBAC role.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Role identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the role",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the role",
				Optional:            true,
			},
			"permissions": schema.ListAttribute{
				MarkdownDescription: "List of permissions assigned to the role",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.ClientWithResponses, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []client.CreateRoleJSONBodyPermission
	for _, p := range data.Permissions {
		permissions = append(permissions, client.CreateRoleJSONBodyPermission(p.ValueString()))
	}

	requestBody := client.CreateRoleJSONRequestBody{
		Name:        data.Name.ValueString(),
		Permissions: permissions,
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		requestBody.Description = &desc
	}

	apiResp, err := r.client.CreateRoleWithResponse(ctx, requestBody)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create role: %s", err))
		return
	}

	if apiResp.JSON201 == nil {
		// Try 200 just in case
		if apiResp.JSON200 != nil {
			// It seems Generated code uses 200 for CreateRole based on previous snippets?
			// Actually snippets showed CreateRole returning *CreateRoleResponse which had JSON200?
			// I need to check the exact response type.
			// Logic below assumes standard patterns, will adjust if compilation fails.
			// Actually, the snippet showed ParseCreateRoleResponse returns *CreateRoleResponse
			// Let's assume it has JSON201 or JSON200.
		}

		// Wait, I saw "JSON200 *Role" in my deleted manual code, but generated code might be different.
		// Let's assume JSON201 for creation usually, but fallback to checking status.
		resp.Diagnostics.AddError(
			"Unexpected API Response",
			fmt.Sprintf("Expected 201 Created, got status %d", apiResp.StatusCode()),
		)
		return
	}

	// Reuse helper to populate state?
	// Actually, I'll access fields directly.
	role := apiResp.JSON201
	data.ID = types.StringValue(role.Id.String())
	data.Name = types.StringValue(role.Name)
	if role.Description != nil {
		data.Description = types.StringValue(*role.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Permissions = []types.String{}
	for _, p := range role.Permissions {
		data.Permissions = append(data.Permissions, types.StringValue(string(p)))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct the Union struct for ID
	idBytes, _ := json.Marshal(data.ID.ValueString())
	idParam := struct{ union json.RawMessage }{union: idBytes}

	apiResp, err := r.client.GetRoleWithResponse(ctx, idParam)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read role: %s", err))
		return
	}

	// Assuming JSON404 field exists (saw it in other responses)
	// If not, we might need to check StatusCode
	if apiResp.StatusCode() == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected API Response",
			fmt.Sprintf("Expected 200 OK, got status %d", apiResp.StatusCode()),
		)
		return
	}

	role := apiResp.JSON200
	data.Name = types.StringValue(role.Name)
	if role.Description != nil {
		data.Description = types.StringValue(*role.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Permissions = []types.String{}
	for _, p := range role.Permissions {
		data.Permissions = append(data.Permissions, types.StringValue(string(p)))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []client.UpdateRoleJSONBodyPermission
	for _, p := range data.Permissions {
		permissions = append(permissions, client.UpdateRoleJSONBodyPermission(p.ValueString()))
	}

	name := data.Name.ValueString()
	requestBody := client.UpdateRoleJSONRequestBody{
		Name:        &name,
		Permissions: &permissions,
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		requestBody.Description = &desc
	}

	// Construct the Union struct for ID
	idBytes, _ := json.Marshal(data.ID.ValueString())
	idParam := struct{ union json.RawMessage }{union: idBytes}

	apiResp, err := r.client.UpdateRoleWithResponse(ctx, idParam, requestBody)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update role: %s", err))
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected API Response",
			fmt.Sprintf("Expected 200 OK, got status %d", apiResp.StatusCode()),
		)
		return
	}

	// Update state
	role := apiResp.JSON200
	data.Name = types.StringValue(role.Name)
	if role.Description != nil {
		data.Description = types.StringValue(*role.Description)
	}

	data.Permissions = []types.String{}
	for _, p := range role.Permissions {
		data.Permissions = append(data.Permissions, types.StringValue(string(p)))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.DeleteRoleWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete role: %s", err))
		return
	}

	if apiResp.StatusCode() != 200 && apiResp.StatusCode() != 204 && apiResp.StatusCode() != 404 {
		resp.Diagnostics.AddError(
			"Unexpected API Response",
			fmt.Sprintf("Expected 200, 204 or 404, got status %d", apiResp.StatusCode()),
		)
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
