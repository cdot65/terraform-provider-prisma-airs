package provider

import (
	"context"
	"encoding/json"

	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &redTeamQuotaDataSource{}

func NewRedTeamQuotaDataSource() datasource.DataSource {
	return &redTeamQuotaDataSource{}
}

type redTeamQuotaDataSource struct {
	client *redteam.Client
}

type RedTeamQuotaDataSourceModel struct {
	Details types.String `tfsdk:"details"`
}

func (d *redTeamQuotaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_quota"
}

func (d *redTeamQuotaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches Red Team quota information.",
		Attributes: map[string]schema.Attribute{
			"details": schema.StringAttribute{
				Computed:    true,
				Description: "Quota details as JSON.",
			},
		},
	}
}

func (d *redTeamQuotaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getRedTeamClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = client
}

func (d *redTeamQuotaDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	quota, err := d.client.GetQuota(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get red team quota", err.Error())
		return
	}

	var state RedTeamQuotaDataSourceModel
	if quota.Details != nil {
		if b, err := json.Marshal(quota.Details); err == nil {
			state.Details = types.StringValue(string(b))
		}
	} else {
		state.Details = types.StringValue("{}")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
