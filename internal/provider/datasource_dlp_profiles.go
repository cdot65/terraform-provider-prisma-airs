package provider

import (
	"context"
	"encoding/json"

	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &dlpProfilesDataSource{}

func NewDlpProfilesDataSource() datasource.DataSource {
	return &dlpProfilesDataSource{}
}

type dlpProfilesDataSource struct {
	client *management.Client
}

type DlpProfilesDataSourceModel struct {
	Limit      types.Int64           `tfsdk:"limit"`
	Offset     types.Int64           `tfsdk:"offset"`
	Items      []DlpProfileItemModel `tfsdk:"items"`
	TotalCount types.Int64           `tfsdk:"total_count"`
}

type DlpProfileItemModel struct {
	ProfileID   types.String `tfsdk:"profile_id"`
	ProfileName types.String `tfsdk:"profile_name"`
	Details     types.String `tfsdk:"details"`
}

func (d *dlpProfilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dlp_profiles"
}

func (d *dlpProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches DLP profiles.",
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
				Description: "Total number of DLP profiles.",
			},
			"items": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of DLP profiles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"profile_id": schema.StringAttribute{
							Computed:    true,
							Description: "DLP profile ID.",
						},
						"profile_name": schema.StringAttribute{
							Computed:    true,
							Description: "DLP profile name.",
						},
						"details": schema.StringAttribute{
							Computed:    true,
							Description: "Profile details as JSON string.",
						},
					},
				},
			},
		},
	}
}

func (d *dlpProfilesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dlpProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DlpProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := management.ListOpts{}
	if !config.Limit.IsNull() && !config.Limit.IsUnknown() {
		opts.Limit = int(config.Limit.ValueInt64())
	}
	if !config.Offset.IsNull() && !config.Offset.IsUnknown() {
		opts.Offset = int(config.Offset.ValueInt64())
	}

	listResp, err := d.client.DlpProfiles.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list DLP profiles", err.Error())
		return
	}

	config.TotalCount = types.Int64Value(int64(len(listResp.Items)))
	config.Items = make([]DlpProfileItemModel, len(listResp.Items))
	for i, item := range listResp.Items {
		// Serialize the full profile as JSON details
		detailsJSON := ""
		b, err := json.Marshal(item)
		if err == nil {
			detailsJSON = string(b)
		}
		config.Items[i] = DlpProfileItemModel{
			ProfileID:   types.StringValue(item.ID),
			ProfileName: types.StringValue(item.Name),
			Details:     types.StringValue(detailsJSON),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
