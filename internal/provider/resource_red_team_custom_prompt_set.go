package provider

import (
	"context"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &redTeamCustomPromptSetResource{}
	_ resource.ResourceWithImportState = &redTeamCustomPromptSetResource{}
)

func NewRedTeamCustomPromptSetResource() resource.Resource {
	return &redTeamCustomPromptSetResource{}
}

type redTeamCustomPromptSetResource struct {
	client *redteam.Client
}

type RedTeamCustomPromptSetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	Active      types.Bool   `tfsdk:"active"`
	Archive     types.Bool   `tfsdk:"archive"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *redTeamCustomPromptSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_custom_prompt_set"
}

func (r *redTeamCustomPromptSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Red Team custom prompt set.",
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
				Description: "The unique identifier of the prompt set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the prompt set.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description of the prompt set. Clearing a nonempty value requires replacement; the old set is archived.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = !req.PlanValue.IsUnknown() && req.PlanValue.ValueString() == "" && req.StateValue.ValueString() != ""
				}, "Replace when clearing a description.", "Replace when clearing a description.")},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Prompt set status.",
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the prompt set is active.",
			},
			"archive": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the prompt set is archived.",
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

func (r *redTeamCustomPromptSetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *redTeamCustomPromptSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedTeamCustomPromptSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := redteam.CustomPromptSetCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	promptSet, err := r.client.CustomAttacks.CreatePromptSet(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create custom prompt set", err.Error())
		return
	}

	mapPromptSetToState(promptSet, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redTeamCustomPromptSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedTeamCustomPromptSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	promptSet, err := r.client.CustomAttacks.GetPromptSet(ctx, state.UUID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read custom prompt set", err.Error())
		return
	}

	// Destroy archives prompt sets (no delete endpoint). An archived set is gone
	// as far as Terraform is concerned, including one archived out-of-band.
	if promptSet.Archive {
		resp.State.RemoveResource(ctx)
		return
	}

	mapPromptSetToState(promptSet, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *redTeamCustomPromptSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RedTeamCustomPromptSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RedTeamCustomPromptSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := rtschema.CustomPromptSetUpdateRequest{
		Name:        aisec.Value(plan.Name.ValueString()),
		Description: aisec.Value(plan.Description.ValueString()),
	}

	_, err := r.client.CustomAttacks.UpdatePromptSetDetails(ctx, state.UUID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update custom prompt set", err.Error())
		return
	}

	promptSet, err := r.client.CustomAttacks.GetPromptSet(ctx, state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated prompt set", err.Error())
		return
	}
	mapPromptSetToState(promptSet, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redTeamCustomPromptSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedTeamCustomPromptSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Archive the prompt set (no delete API)
	_, err := r.client.CustomAttacks.ArchivePromptSet(ctx, state.UUID.ValueString(), redteam.CustomPromptSetArchiveRequest{
		Archive: true,
	})
	finishDelete(ctx, err, "custom prompt set (archive)", func(ctx context.Context) (bool, error) {
		ps, getErr := r.client.CustomAttacks.GetPromptSet(ctx, state.UUID.ValueString())
		if isNotFound(getErr) {
			return true, nil
		}
		if getErr != nil {
			return false, getErr
		}
		return ps.Archive, nil
	}, &resp.Diagnostics)
}

func (r *redTeamCustomPromptSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	uuid := req.ID

	promptSet, err := r.client.CustomAttacks.GetPromptSet(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import prompt set", err.Error())
		return
	}

	if promptSet.Archive {
		resp.Diagnostics.AddError("Prompt set is archived", "An archived prompt set cannot be imported as a managed active set.")
		return
	}
	var state RedTeamCustomPromptSetResourceModel
	mapPromptSetToState(promptSet, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapPromptSetToState(ps *redteam.CustomPromptSetResponse, state *RedTeamCustomPromptSetResourceModel) {
	state.ID = types.StringValue(ps.UUID)
	state.UUID = types.StringValue(ps.UUID)
	state.Name = types.StringValue(ps.Name)
	state.Description = types.StringValue(ps.Description)
	state.Status = types.StringValue(ps.Status)
	state.Active = types.BoolValue(ps.Active)
	state.Archive = types.BoolValue(ps.Archive)
	state.CreatedAt = types.StringValue(ps.CreatedAt)
	state.UpdatedAt = types.StringValue(ps.UpdatedAt)
}
