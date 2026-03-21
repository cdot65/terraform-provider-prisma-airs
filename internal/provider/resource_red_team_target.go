package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &redTeamTargetResource{}
	_ resource.ResourceWithImportState = &redTeamTargetResource{}
)

func NewRedTeamTargetResource() resource.Resource {
	return &redTeamTargetResource{}
}

type redTeamTargetResource struct {
	client *redteam.Client
}

type RedTeamTargetResourceModel struct {
	ID               types.String `tfsdk:"id"`
	UUID             types.String `tfsdk:"uuid"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	TargetType       types.String `tfsdk:"target_type"`
	ConnectionType   types.String `tfsdk:"connection_type"`
	ConnectionParams types.String `tfsdk:"connection_params"`
	Status           types.String `tfsdk:"status"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *redTeamTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_target"
}

func (r *redTeamTargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Red Team target.",
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
				Description: "The unique identifier of the target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the target.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the target.",
			},
			"target_type": schema.StringAttribute{
				Optional:    true,
				Description: "Target type (APPLICATION, AGENT, MODEL).",
			},
			"connection_type": schema.StringAttribute{
				Optional:    true,
				Description: "Connection type (DATABRICKS, BEDROCK, OPENAI, HUGGING_FACE, CUSTOM, REST, STREAMING).",
			},
			"connection_params": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Connection parameters as a JSON string.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Target status.",
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

func (r *redTeamTargetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getRedTeamClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = client
}

func (r *redTeamTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := redteam.TargetCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if !plan.TargetType.IsNull() && !plan.TargetType.IsUnknown() {
		createReq.TargetType = redteam.TargetType(plan.TargetType.ValueString())
	}
	if !plan.ConnectionType.IsNull() && !plan.ConnectionType.IsUnknown() {
		createReq.ConnectionType = redteam.TargetConnectionType(plan.ConnectionType.ValueString())
	}
	if !plan.ConnectionParams.IsNull() && !plan.ConnectionParams.IsUnknown() && plan.ConnectionParams.ValueString() != "" {
		var params map[string]any
		if err := json.Unmarshal([]byte(plan.ConnectionParams.ValueString()), &params); err != nil {
			resp.Diagnostics.AddError("Invalid connection_params JSON", err.Error())
			return
		}
		createReq.ConnectionParams = params
	}

	target, err := r.client.Targets.Create(ctx, createReq, false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create red team target", err.Error())
		return
	}

	mapTargetToState(target, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redTeamTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target, err := r.client.Targets.Get(ctx, state.UUID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read red team target", err.Error())
		return
	}

	mapTargetToState(target, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *redTeamTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := redteam.TargetUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if !plan.TargetType.IsNull() && !plan.TargetType.IsUnknown() {
		updateReq.TargetType = redteam.TargetType(plan.TargetType.ValueString())
	}
	if !plan.ConnectionType.IsNull() && !plan.ConnectionType.IsUnknown() {
		updateReq.ConnectionType = redteam.TargetConnectionType(plan.ConnectionType.ValueString())
	}
	if !plan.ConnectionParams.IsNull() && !plan.ConnectionParams.IsUnknown() && plan.ConnectionParams.ValueString() != "" {
		var params map[string]any
		if err := json.Unmarshal([]byte(plan.ConnectionParams.ValueString()), &params); err != nil {
			resp.Diagnostics.AddError("Invalid connection_params JSON", err.Error())
			return
		}
		updateReq.ConnectionParams = params
	}

	target, err := r.client.Targets.Update(ctx, state.UUID.ValueString(), updateReq, false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update red team target", err.Error())
		return
	}

	mapTargetToState(target, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redTeamTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Targets.Delete(ctx, state.UUID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "failed to parse response JSON") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete red team target", err.Error())
		return
	}
}

func (r *redTeamTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	uuid := req.ID

	target, err := r.client.Targets.Get(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Target not found", "No target with UUID: "+uuid)
		return
	}

	var state RedTeamTargetResourceModel
	mapTargetToState(target, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapTargetToState(target *redteam.TargetResponse, state *RedTeamTargetResourceModel) {
	state.ID = types.StringValue(target.UUID)
	state.UUID = types.StringValue(target.UUID)
	state.Name = types.StringValue(target.Name)
	state.Description = types.StringValue(target.Description)
	state.Status = types.StringValue(string(target.Status))
	state.CreatedAt = types.StringValue(target.CreatedAt)
	state.UpdatedAt = types.StringValue(target.UpdatedAt)

	if string(target.TargetType) != "" {
		state.TargetType = types.StringValue(string(target.TargetType))
	}
	if string(target.ConnectionType) != "" {
		state.ConnectionType = types.StringValue(string(target.ConnectionType))
	}
	if target.ConnectionParams != nil {
		paramsJSON, err := json.Marshal(target.ConnectionParams)
		if err == nil {
			state.ConnectionParams = types.StringValue(string(paramsJSON))
		}
	}
}

func getRedTeamClient(data any) (*redteam.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	pd, ok := data.(*ProviderData)
	if !ok || pd == nil {
		diags.AddError("Missing provider config", "Provider not configured")
		return nil, diags
	}
	if pd.RedTeamClient == nil {
		diags.AddError("Red team client not configured", "OAuth2 credentials required")
		return nil, diags
	}
	return pd.RedTeamClient, diags
}
