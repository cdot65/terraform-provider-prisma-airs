package provider

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = &modelScanResource{}
	_ resource.ResourceWithImportState = &modelScanResource{}
)

func NewModelScanResource() resource.Resource {
	return &modelScanResource{}
}

type modelScanResource struct {
	client *modelsecurity.Client
}

type ModelScanResourceModel struct {
	ID                types.String `tfsdk:"id"`
	UUID              types.String `tfsdk:"uuid"`
	ModelURI          types.String `tfsdk:"model_uri"`
	ScanOrigin        types.String `tfsdk:"scan_origin"`
	SourceType        types.String `tfsdk:"source_type"`
	SecurityGroupUUID types.String `tfsdk:"security_group_uuid"`
	Labels            types.Map    `tfsdk:"labels"`
	EvalOutcome       types.String `tfsdk:"eval_outcome"`
	EvalSummary       types.String `tfsdk:"eval_summary"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func (r *modelScanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_scan"
}

func (r *modelScanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Model Security scan.",
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
				Description: "The unique identifier of the scan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model_uri": schema.StringAttribute{
				Required:    true,
				Description: "URI of the model to scan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scan_origin": schema.StringAttribute{
				Required:    true,
				Description: "Scan origin (MODEL_SECURITY_SDK, HUGGING_FACE).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_type": schema.StringAttribute{
				Computed:    true,
				Description: "Source type (LOCAL, HUGGING_FACE, S3, GCS, AZURE, ARTIFACTORY, GITLAB, ALL).",
			},
			"security_group_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the security group to use for this scan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Key-value labels for the scan.",
			},
			"eval_outcome": schema.StringAttribute{
				Computed:    true,
				Description: "Evaluation outcome (PENDING, ALLOWED, BLOCKED, ERROR).",
			},
			"eval_summary": schema.StringAttribute{
				Computed:    true,
				Description: "Evaluation summary as JSON.",
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

func (r *modelScanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *modelScanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModelScanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := modelsecurity.ScanCreateRequest{
		ModelURI:   plan.ModelURI.ValueString(),
		ScanOrigin: modelsecurity.ScanOrigin(plan.ScanOrigin.ValueString()),
	}

	if !plan.SecurityGroupUUID.IsNull() && !plan.SecurityGroupUUID.IsUnknown() {
		createReq.SecurityGroupUUID = plan.SecurityGroupUUID.ValueString()
	}

	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		labelsMap := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labelsMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Labels = mapToLabels(labelsMap)
	}

	scan, err := r.client.Scans.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create model scan", err.Error())
		return
	}

	mapScanToState(ctx, scan, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *modelScanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModelScanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scan, err := r.client.Scans.Get(ctx, state.UUID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read model scan", err.Error())
		return
	}

	mapScanToState(ctx, scan, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *modelScanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModelScanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ModelScanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only labels can be updated on a scan.
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		labelsMap := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labelsMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		_, err := r.client.Scans.SetLabels(ctx, state.UUID.ValueString(), modelsecurity.LabelsCreateRequest{
			Labels: mapToLabels(labelsMap),
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to update scan labels", err.Error())
			return
		}
	}

	// Re-read the scan.
	scan, err := r.client.Scans.Get(ctx, state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read model scan after update", err.Error())
		return
	}

	mapScanToState(ctx, scan, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *modelScanResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Scans cannot be deleted via API — just remove from state.
	_ = ctx
	_ = resp
}

func (r *modelScanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	uuid := req.ID

	scan, err := r.client.Scans.Get(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Scan not found", "No scan with UUID: "+uuid)
		return
	}

	var state ModelScanResourceModel
	mapScanToState(ctx, scan, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapToLabels converts a map[string]string to []modelsecurity.Label.
func mapToLabels(m map[string]string) []modelsecurity.Label {
	labels := make([]modelsecurity.Label, 0, len(m))
	for k, v := range m {
		labels = append(labels, modelsecurity.Label{Key: k, Value: v})
	}
	return labels
}

// labelsToMap converts []modelsecurity.Label to map[string]string.
func labelsToMap(labels []modelsecurity.Label) map[string]string {
	m := make(map[string]string, len(labels))
	for _, l := range labels {
		m[l.Key] = l.Value
	}
	return m
}

func mapScanToState(ctx context.Context, scan *modelsecurity.ScanBaseResponse, state *ModelScanResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(scan.UUID)
	state.UUID = types.StringValue(scan.UUID)
	state.ModelURI = types.StringValue(scan.ModelURI)
	state.ScanOrigin = types.StringValue(string(scan.ScanOrigin))
	state.SourceType = types.StringValue(string(scan.SourceType))
	state.SecurityGroupUUID = types.StringValue(scan.SecurityGroupUUID)
	state.EvalOutcome = types.StringValue(string(scan.EvalOutcome))
	state.CreatedAt = types.StringValue(scan.CreatedAt)
	state.UpdatedAt = types.StringValue(scan.UpdatedAt)

	if scan.EvalSummary != nil {
		summaryJSON, err := json.Marshal(scan.EvalSummary)
		if err == nil {
			state.EvalSummary = types.StringValue(string(summaryJSON))
		}
	} else {
		state.EvalSummary = types.StringNull()
	}

	if len(scan.Labels) > 0 {
		labelsMap, d := types.MapValueFrom(ctx, types.StringType, labelsToMap(scan.Labels))
		diags.Append(d...)
		state.Labels = labelsMap
	} else {
		state.Labels = types.MapNull(types.StringType)
	}
}
