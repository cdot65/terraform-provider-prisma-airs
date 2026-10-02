package provider

import (
	"context"
	"os"

	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/cdot65/prisma-airs-provider/internal/products"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &PrismaAIRSProvider{}

type PrismaAIRSProvider struct{ version string }

type PrismaAIRSProviderModel struct {
	ClientID      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	TsgID         types.String `tfsdk:"tsg_id"`
	TokenEndpoint types.String `tfsdk:"token_endpoint"`
	Runtime       types.Object `tfsdk:"runtime"`
	RedTeam       types.Object `tfsdk:"red_team"`
	SupplyChain   types.Object `tfsdk:"supply_chain"`
}

func (p *PrismaAIRSProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "prisma-airs"
	resp.Version = p.version
}

func (p *PrismaAIRSProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Prisma AIRS: AI Runtime Security, AI Red Teaming, and AI Supply Chain Security. AI Gateway functionality is not yet implemented.",
		Attributes: map[string]schema.Attribute{
			"client_id":      schema.StringAttribute{Description: "Shared OAuth2 client ID. Can also be set via PANW_MGMT_CLIENT_ID.", Optional: true},
			"client_secret":  schema.StringAttribute{Description: "Shared OAuth2 client secret. Can also be set via PANW_MGMT_CLIENT_SECRET.", Optional: true, Sensitive: true},
			"tsg_id":         schema.StringAttribute{Description: "Tenant Service Group ID. Can also be set via PANW_MGMT_TSG_ID.", Optional: true},
			"token_endpoint": schema.StringAttribute{Description: "Shared OAuth2 token endpoint override. Can also be set via PANW_MGMT_TOKEN_ENDPOINT.", Optional: true},
		},
		Blocks: map[string]schema.Block{},
	}
	for _, definition := range products.All() {
		if !definition.Implemented {
			continue
		}
		attributes := map[string]schema.Attribute{}
		for _, endpoint := range definition.Endpoints {
			attributes[endpoint.Name] = schema.StringAttribute{Description: endpoint.Description + " Can also be set via " + endpoint.Environment + ".", Optional: true}
		}
		resp.Schema.Blocks[definition.ID] = schema.SingleNestedBlock{Description: definition.Label + " endpoint settings. Omit to use environment variables or SDK defaults.", Attributes: attributes}
	}
}

func (p *PrismaAIRSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PrismaAIRSProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	credentials := product.Credentials{
		ClientID:      stringValueOrEnv(config.ClientID, "PANW_MGMT_CLIENT_ID"),
		ClientSecret:  stringValueOrEnv(config.ClientSecret, "PANW_MGMT_CLIENT_SECRET"),
		TsgID:         stringValueOrEnv(config.TsgID, "PANW_MGMT_TSG_ID"),
		TokenEndpoint: stringValueOrEnv(config.TokenEndpoint, "PANW_MGMT_TOKEN_ENDPOINT"),
	}
	data := product.Data{}
	// Client construction is local and lazy: API authorization is checked only
	// when a resource uses its product. Missing credentials retain that behavior.
	if credentials.Complete() {
		settings := map[string]types.Object{"runtime": config.Runtime, "red_team": config.RedTeam, "supply_chain": config.SupplyChain}
		for _, definition := range products.All() {
			if !definition.Implemented {
				continue
			}
			endpoints := map[string]string{}
			attributes := settings[definition.ID].Attributes()
			for _, endpoint := range definition.Endpoints {
				value := types.StringNull()
				if configured, ok := attributes[endpoint.Name].(types.String); ok {
					value = configured
				}
				endpoints[endpoint.Name] = stringValueOrEnv(value, endpoint.Environment)
			}
			client, err := definition.Configure(credentials, endpoints)
			if err != nil {
				resp.Diagnostics.AddError("Failed to configure "+definition.Label, err.Error())
				return
			}
			data[definition.ID] = client
		}
	}
	resp.DataSourceData = data
	resp.ResourceData = data
}

func (p *PrismaAIRSProvider) Resources(_ context.Context) []func() resource.Resource {
	var constructors []func() resource.Resource
	for _, definition := range products.All() {
		for _, entry := range definition.Resources {
			constructors = append(constructors, entry.New)
		}
	}
	return constructors
}

func (p *PrismaAIRSProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	var constructors []func() datasource.DataSource
	for _, definition := range products.All() {
		for _, entry := range definition.DataSources {
			constructors = append(constructors, entry.New)
		}
	}
	return constructors
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &PrismaAIRSProvider{version: version} }
}

func stringValueOrEnv(val types.String, envKey string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envKey)
}
