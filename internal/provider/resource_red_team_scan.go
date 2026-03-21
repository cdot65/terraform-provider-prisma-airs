package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &redTeamScanResource{}
	_ resource.ResourceWithImportState = &redTeamScanResource{}
)

func NewRedTeamScanResource() resource.Resource {
	return &redTeamScanResource{}
}

type redTeamScanResource struct {
	client *redteam.Client
}

type RedTeamScanResourceModel struct {
	ID         types.String `tfsdk:"id"`
	JobID      types.String `tfsdk:"job_id"`
	Name       types.String `tfsdk:"name"`
	TargetID   types.String `tfsdk:"target_id"`
	JobType    types.String `tfsdk:"job_type"`
	Status     types.String `tfsdk:"status"`
	Stats      types.String `tfsdk:"stats"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	FinishedAt types.String `tfsdk:"finished_at"`
}

func (r *redTeamScanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_scan"
}

func (r *redTeamScanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Red Team scan job.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform resource ID (same as job_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the scan job.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the scan job.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the target to scan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"job_type": schema.StringAttribute{
				Required:    true,
				Description: "Job type (STATIC, DYNAMIC, CUSTOM).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Job status.",
			},
			"stats": schema.StringAttribute{
				Computed:    true,
				Description: "Job statistics as JSON.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp.",
			},
			"finished_at": schema.StringAttribute{
				Computed:    true,
				Description: "Completion timestamp.",
			},
		},
	}
}

func (r *redTeamScanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *redTeamScanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedTeamScanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := redteam.JobCreateRequest{
		TargetID: plan.TargetID.ValueString(),
		JobType:  redteam.JobType(plan.JobType.ValueString()),
	}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		createReq.Name = plan.Name.ValueString()
	}

	job, err := r.client.Scans.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create red team scan", err.Error())
		return
	}

	mapJobToState(job, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redTeamScanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedTeamScanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.Scans.Get(ctx, state.JobID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read red team scan", err.Error())
		return
	}

	mapJobToState(job, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *redTeamScanResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Red team scans cannot be updated; all mutable fields force replacement.")
}

func (r *redTeamScanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedTeamScanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Abort running scans on destroy
	status := state.Status.ValueString()
	if status == string(redteam.JobStatusRunning) || status == string(redteam.JobStatusQueued) || status == string(redteam.JobStatusInit) {
		_, _ = r.client.Scans.Abort(ctx, state.JobID.ValueString())
	}
	// Scans cannot be deleted via API — just remove from state.
}

func (r *redTeamScanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	jobID := req.ID

	job, err := r.client.Scans.Get(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError("Scan not found", "No scan with ID: "+jobID)
		return
	}

	var state RedTeamScanResourceModel
	mapJobToState(job, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapJobToState(job *redteam.JobResponse, state *RedTeamScanResourceModel) {
	state.ID = types.StringValue(job.ID)
	state.JobID = types.StringValue(job.ID)
	state.Name = types.StringValue(job.Name)
	state.TargetID = types.StringValue(job.TargetID)
	state.JobType = types.StringValue(string(job.JobType))
	state.Status = types.StringValue(string(job.Status))
	state.CreatedAt = types.StringValue(job.CreatedAt)
	state.UpdatedAt = types.StringValue(job.UpdatedAt)
	state.FinishedAt = types.StringValue(job.FinishedAt)

	if job.Stats != nil {
		statsJSON, err := json.Marshal(job.Stats)
		if err == nil {
			state.Stats = types.StringValue(string(statsJSON))
		}
	} else {
		state.Stats = types.StringNull()
	}
}
