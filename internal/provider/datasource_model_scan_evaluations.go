package provider

import (
	"context"
	"encoding/json"

	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &modelScanEvaluationsDataSource{}

func NewModelScanEvaluationsDataSource() datasource.DataSource {
	return &modelScanEvaluationsDataSource{}
}

type modelScanEvaluationsDataSource struct {
	client *modelsecurity.Client
}

type ModelScanEvaluationsDataSourceModel struct {
	ScanUUID    types.String               `tfsdk:"scan_uuid"`
	Evaluations []ModelScanEvaluationModel `tfsdk:"evaluations"`
}

type ModelScanEvaluationModel struct {
	UUID             types.String `tfsdk:"uuid"`
	ScanUUID         types.String `tfsdk:"scan_uuid"`
	RuleInstanceUUID types.String `tfsdk:"rule_instance_uuid"`
	RuleName         types.String `tfsdk:"rule_name"`
	Result           types.String `tfsdk:"result"`
	Details          types.String `tfsdk:"details"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (d *modelScanEvaluationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_scan_evaluations"
}

func (d *modelScanEvaluationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches evaluations for a Model Security scan.",
		Attributes: map[string]schema.Attribute{
			"scan_uuid": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the scan to fetch evaluations for.",
			},
			"evaluations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of rule evaluations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "Evaluation UUID.",
						},
						"scan_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "Scan UUID.",
						},
						"rule_instance_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "Rule instance UUID.",
						},
						"rule_name": schema.StringAttribute{
							Computed:    true,
							Description: "Rule name.",
						},
						"result": schema.StringAttribute{
							Computed:    true,
							Description: "Evaluation result (PASSED, FAILED, ERROR).",
						},
						"details": schema.StringAttribute{
							Computed:    true,
							Description: "Evaluation details as JSON.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Creation timestamp.",
						},
					},
				},
			},
		},
	}
}

func (d *modelScanEvaluationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getModelSecClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = client
}

func (d *modelScanEvaluationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ModelScanEvaluationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	evalsResp, err := d.client.Scans.GetEvaluations(ctx, config.ScanUUID.ValueString(), modelsecurity.EvaluationListOpts{Limit: 100})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list scan evaluations", err.Error())
		return
	}

	var state ModelScanEvaluationsDataSourceModel
	state.ScanUUID = config.ScanUUID

	for _, item := range evalsResp.Items {
		detailsJSON := ""
		if b, err := json.Marshal(item); err == nil {
			detailsJSON = string(b)
		}
		state.Evaluations = append(state.Evaluations, ModelScanEvaluationModel{
			UUID:             types.StringValue(item.UUID),
			ScanUUID:         types.StringValue(item.ScanUUID),
			RuleInstanceUUID: types.StringValue(item.RuleInstanceUUID),
			RuleName:         types.StringValue(item.RuleName),
			Result:           types.StringValue(string(item.Result)),
			Details:          types.StringValue(detailsJSON),
			CreatedAt:        types.StringValue(item.CreatedAt),
		})
	}

	if state.Evaluations == nil {
		state.Evaluations = []ModelScanEvaluationModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
