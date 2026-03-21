package provider

import (
	"context"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &modelSecurityGroupResource{}
	_ resource.ResourceWithImportState = &modelSecurityGroupResource{}
)

func NewModelSecurityGroupResource() resource.Resource {
	return &modelSecurityGroupResource{}
}

type modelSecurityGroupResource struct {
	client *modelsecurity.Client
}

type ModelSecurityGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	SourceType  types.String `tfsdk:"source_type"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *modelSecurityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_security_group"
}

func (r *modelSecurityGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Model Security group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform resource ID (same as uuid).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the security group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the security group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the security group.",
			},
			"source_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Source type (LOCAL, HUGGING_FACE, S3, GCS, AZURE, ARTIFACTORY, GITLAB, ALL).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Group state (PENDING, ACTIVE).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp.",
			},
		},
	}
}

func (r *modelSecurityGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getModelSecClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = client
}

func (r *modelSecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModelSecurityGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := modelsecurity.ModelSecurityGroupCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if !plan.SourceType.IsNull() && !plan.SourceType.IsUnknown() {
		createReq.SourceType = modelsecurity.SourceType(plan.SourceType.ValueString())
	}

	group, err := r.client.SecurityGroups.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create model security group", err.Error())
		return
	}

	mapSecurityGroupToState(group, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *modelSecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModelSecurityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.SecurityGroups.Get(ctx, state.UUID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read model security group", err.Error())
		return
	}

	mapSecurityGroupToState(group, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *modelSecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModelSecurityGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ModelSecurityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := modelsecurity.ModelSecurityGroupUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	group, err := r.client.SecurityGroups.Update(ctx, state.UUID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update model security group", err.Error())
		return
	}

	mapSecurityGroupToState(group, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *modelSecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModelSecurityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SecurityGroups.Delete(ctx, state.UUID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "failed to parse response JSON") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete model security group", err.Error())
		return
	}
}

func (r *modelSecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	uuid := req.ID

	group, err := r.client.SecurityGroups.Get(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Group not found", "No security group with UUID: "+uuid)
		return
	}

	var state ModelSecurityGroupResourceModel
	mapSecurityGroupToState(group, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapSecurityGroupToState(group *modelsecurity.ModelSecurityGroupResponse, state *ModelSecurityGroupResourceModel) {
	state.ID = types.StringValue(group.UUID)
	state.UUID = types.StringValue(group.UUID)
	state.Name = types.StringValue(group.Name)
	state.Description = types.StringValue(group.Description)
	state.SourceType = types.StringValue(string(group.SourceType))
	state.State = types.StringValue(string(group.State))
	state.CreatedAt = types.StringValue(group.CreatedAt)
	state.UpdatedAt = types.StringValue(group.UpdatedAt)
}

func getModelSecClient(data any) (*modelsecurity.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	pd, ok := data.(*ProviderData)
	if !ok || pd == nil {
		diags.AddError("Missing provider config", "Provider not configured")
		return nil, diags
	}
	if pd.ModelSecClient == nil {
		diags.AddError("Model security client not configured", "OAuth2 credentials required")
		return nil, diags
	}
	return pd.ModelSecClient, diags
}
