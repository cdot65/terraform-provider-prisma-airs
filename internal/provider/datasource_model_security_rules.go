package provider

import (
	"context"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &modelSecurityRulesDataSource{}

func NewModelSecurityRulesDataSource() datasource.DataSource {
	return &modelSecurityRulesDataSource{}
}

type modelSecurityRulesDataSource struct {
	client *modelsecurity.Client
}

type ModelSecurityRulesDataSourceModel struct {
	Rules []ModelSecurityRuleModel `tfsdk:"rules"`
}

type ModelSecurityRuleModel struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	SourceType  types.String `tfsdk:"source_type"`
	RuleType    types.String `tfsdk:"rule_type"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (d *modelSecurityRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_security_rules"
}

func (d *modelSecurityRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the catalog of Model Security rules.",
		Attributes: map[string]schema.Attribute{
			"rules": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of security rules.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "Rule UUID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Rule name.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Rule description.",
						},
						"source_type": schema.StringAttribute{
							Computed:    true,
							Description: "Source type.",
						},
						"rule_type": schema.StringAttribute{
							Computed:    true,
							Description: "Rule type (METADATA, ARTIFACT).",
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

func (d *modelSecurityRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *modelSecurityRulesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	rulesResp, err := d.client.SecurityRules.List(ctx, modelsecurity.RuleListOpts{Limit: 100})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list model security rules", err.Error())
		return
	}

	var state ModelSecurityRulesDataSourceModel
	for _, item := range rulesResp.Items {
		// CompatibleSources is a slice; join for display
		sourceTypes := ""
		if len(item.CompatibleSources) > 0 {
			parts := make([]string, len(item.CompatibleSources))
			for j, s := range item.CompatibleSources {
				parts[j] = string(s)
			}
			sourceTypes = strings.Join(parts, ",")
		}
		state.Rules = append(state.Rules, ModelSecurityRuleModel{
			UUID:        types.StringValue(item.UUID),
			Name:        types.StringValue(item.Name),
			Description: types.StringValue(item.Description),
			SourceType:  types.StringValue(sourceTypes),
			RuleType:    types.StringValue(string(item.RuleType)),
			CreatedAt:   types.StringValue(""),
		})
	}

	if state.Rules == nil {
		state.Rules = []ModelSecurityRuleModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
