package provider

import (
	"context"
	"encoding/json"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &deploymentProfilesDataSource{}

func NewDeploymentProfilesDataSource() datasource.DataSource {
	return &deploymentProfilesDataSource{}
}

type deploymentProfilesDataSource struct {
	client *airsruntime.Client
}

type DeploymentProfilesDataSourceModel struct {
	Limit      types.Int64                  `tfsdk:"limit"`
	Offset     types.Int64                  `tfsdk:"offset"`
	Items      []DeploymentProfileItemModel `tfsdk:"items"`
	TotalCount types.Int64                  `tfsdk:"total_count"`
}

type DeploymentProfileItemModel struct {
	ProfileID   types.String `tfsdk:"profile_id"`
	ProfileName types.String `tfsdk:"profile_name"`
	AuthCode    types.String `tfsdk:"auth_code"`
	Details     types.String `tfsdk:"details"`
}

func (d *deploymentProfilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_profiles"
}

func (d *deploymentProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches deployment profiles.",
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
				Description: "Number of deployment profiles returned.",
			},
			"items": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of deployment profiles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"profile_id": schema.StringAttribute{
							Computed:    true,
							Description: "Deployment profile ID.",
						},
						"profile_name": schema.StringAttribute{
							Computed:    true,
							Description: "Deployment profile name.",
						},
						"auth_code": schema.StringAttribute{
							Computed:    true,
							Description: "Auth code for API key creation.",
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

func (d *deploymentProfilesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deploymentProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DeploymentProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := airsruntime.ListOpts{}
	if !config.Limit.IsNull() && !config.Limit.IsUnknown() {
		opts.Limit = int(config.Limit.ValueInt64())
	}
	if !config.Offset.IsNull() && !config.Offset.IsUnknown() {
		opts.Offset = int(config.Offset.ValueInt64())
	}

	listResp, err := d.client.DeploymentProfiles.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list deployment profiles", err.Error())
		return
	}

	config.TotalCount = types.Int64Value(int64(len(listResp.Items)))
	config.Items = make([]DeploymentProfileItemModel, len(listResp.Items))
	for i, item := range listResp.Items {
		// Serialize the full profile as JSON details
		detailsJSON := ""
		b, err := json.Marshal(item)
		if err == nil {
			detailsJSON = string(b)
		}
		config.Items[i] = DeploymentProfileItemModel{
			ProfileID:   types.StringValue(item.AuthCode),
			ProfileName: types.StringValue(item.DpName),
			AuthCode:    types.StringValue(item.AuthCode),
			Details:     types.StringValue(detailsJSON),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
