package provider

import (
	"context"
	"encoding/json"

	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &scanLogsDataSource{}

func NewScanLogsDataSource() datasource.DataSource {
	return &scanLogsDataSource{}
}

type scanLogsDataSource struct {
	client *management.Client
}

type ScanLogsDataSourceModel struct {
	Limit      types.Int64        `tfsdk:"limit"`
	Offset     types.Int64        `tfsdk:"offset"`
	Items      []ScanLogItemModel `tfsdk:"items"`
	TotalCount types.Int64        `tfsdk:"total_count"`
}

type ScanLogItemModel struct {
	LogID     types.String `tfsdk:"log_id"`
	Details   types.String `tfsdk:"details"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *scanLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scan_logs"
}

func (d *scanLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches scan activity logs.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of results to return.",
			},
			"offset": schema.Int64Attribute{
				Optional:    true,
				Description: "Offset for pagination.",
			},
			"total_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Total number of scan logs.",
			},
			"items": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of scan log entries.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"log_id": schema.StringAttribute{
							Computed:    true,
							Description: "Scan log ID.",
						},
						"details": schema.StringAttribute{
							Computed:    true,
							Description: "Log details as JSON string.",
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

func (d *scanLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getMgmtClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = client
}

func (d *scanLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ScanLogsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := management.ScanLogListOpts{}
	if !config.Limit.IsNull() && !config.Limit.IsUnknown() {
		opts.Limit = int(config.Limit.ValueInt64())
	}
	if !config.Offset.IsNull() && !config.Offset.IsUnknown() {
		opts.Offset = int(config.Offset.ValueInt64())
	}

	listResp, err := d.client.ScanLogs.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list scan logs", err.Error())
		return
	}

	config.TotalCount = types.Int64Value(int64(listResp.TotalCount))
	config.Items = make([]ScanLogItemModel, len(listResp.Items))
	for i, item := range listResp.Items {
		detailsJSON := ""
		if item.Details != nil {
			b, err := json.Marshal(item.Details)
			if err == nil {
				detailsJSON = string(b)
			}
		}
		config.Items[i] = ScanLogItemModel{
			LogID:     types.StringValue(item.LogID),
			Details:   types.StringValue(detailsJSON),
			CreatedAt: types.StringValue(item.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
