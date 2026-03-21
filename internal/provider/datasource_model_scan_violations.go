package provider

import (
	"context"
	"encoding/json"

	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &modelScanViolationsDataSource{}

func NewModelScanViolationsDataSource() datasource.DataSource {
	return &modelScanViolationsDataSource{}
}

type modelScanViolationsDataSource struct {
	client *modelsecurity.Client
}

type ModelScanViolationsDataSourceModel struct {
	ScanUUID   types.String              `tfsdk:"scan_uuid"`
	Violations []ModelScanViolationModel `tfsdk:"violations"`
}

type ModelScanViolationModel struct {
	UUID      types.String `tfsdk:"uuid"`
	ScanUUID  types.String `tfsdk:"scan_uuid"`
	RuleName  types.String `tfsdk:"rule_name"`
	Details   types.String `tfsdk:"details"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *modelScanViolationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_scan_violations"
}

func (d *modelScanViolationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches rule violations for a Model Security scan.",
		Attributes: map[string]schema.Attribute{
			"scan_uuid": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the scan to fetch violations for.",
			},
			"violations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of rule violations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "Violation UUID.",
						},
						"scan_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "Scan UUID.",
						},
						"rule_name": schema.StringAttribute{
							Computed:    true,
							Description: "Rule name.",
						},
						"details": schema.StringAttribute{
							Computed:    true,
							Description: "Violation details as JSON.",
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

func (d *modelScanViolationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *modelScanViolationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ModelScanViolationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	violResp, err := d.client.Scans.GetViolations(ctx, config.ScanUUID.ValueString(), modelsecurity.ViolationListOpts{Limit: 100})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list scan violations", err.Error())
		return
	}

	var state ModelScanViolationsDataSourceModel
	state.ScanUUID = config.ScanUUID

	for _, item := range violResp.Items {
		detailsJSON := ""
		if b, err := json.Marshal(item); err == nil {
			detailsJSON = string(b)
		}
		state.Violations = append(state.Violations, ModelScanViolationModel{
			UUID:      types.StringValue(item.UUID),
			ScanUUID:  config.ScanUUID,
			RuleName:  types.StringValue(item.RuleName),
			Details:   types.StringValue(detailsJSON),
			CreatedAt: types.StringValue(item.CreatedAt),
		})
	}

	if state.Violations == nil {
		state.Violations = []ModelScanViolationModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
