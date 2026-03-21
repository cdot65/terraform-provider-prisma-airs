package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cdot65/prisma-airs-go/aisec/redteam"
)

var _ datasource.DataSource = &redTeamCategoriesDataSource{}

func NewRedTeamCategoriesDataSource() datasource.DataSource {
	return &redTeamCategoriesDataSource{}
}

type redTeamCategoriesDataSource struct {
	client *redteam.Client
}

type RedTeamCategoriesDataSourceModel struct {
	Categories []RedTeamCategoryModel `tfsdk:"categories"`
}

type RedTeamCategoryModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	SubCategories types.String `tfsdk:"sub_categories"`
	Details       types.String `tfsdk:"details"`
}

func (d *redTeamCategoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_categories"
}

func (d *redTeamCategoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches Red Team attack categories.",
		Attributes: map[string]schema.Attribute{
			"categories": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of attack categories.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Category ID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Category name.",
						},
						"sub_categories": schema.StringAttribute{
							Computed:    true,
							Description: "Sub-categories as JSON.",
						},
						"details": schema.StringAttribute{
							Computed:    true,
							Description: "Category details as JSON.",
						},
					},
				},
			},
		},
	}
}

func (d *redTeamCategoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *redTeamCategoriesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	categories, err := d.client.Scans.GetCategories(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list red team categories", err.Error())
		return
	}

	var state RedTeamCategoriesDataSourceModel
	for _, cat := range categories {
		subCatsJSON := ""
		if cat.SubCategories != nil {
			if b, err := json.Marshal(cat.SubCategories); err == nil {
				subCatsJSON = string(b)
			}
		}
		detailsJSON := ""
		if cat.Details != nil {
			if b, err := json.Marshal(cat.Details); err == nil {
				detailsJSON = string(b)
			}
		}
		state.Categories = append(state.Categories, RedTeamCategoryModel{
			ID:            types.StringValue(cat.ID),
			Name:          types.StringValue(cat.Name),
			SubCategories: types.StringValue(subCatsJSON),
			Details:       types.StringValue(detailsJSON),
		})
	}

	if state.Categories == nil {
		state.Categories = []RedTeamCategoryModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
