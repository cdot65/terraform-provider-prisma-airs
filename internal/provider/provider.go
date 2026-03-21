package provider

import (
	"context"
	"os"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/cdot65/prisma-airs-go/aisec/scan"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &PrismaAIRSProvider{}

// PrismaAIRSProvider defines the provider implementation.
type PrismaAIRSProvider struct {
	version string
}

// PrismaAIRSProviderModel describes the provider data model.
type PrismaAIRSProviderModel struct {
	// Scan API (API Key auth)
	APIKey      types.String `tfsdk:"api_key"`
	APIToken    types.String `tfsdk:"api_token"`
	ProfileName types.String `tfsdk:"profile_name"`
	Endpoint    types.String `tfsdk:"endpoint"`

	// Management / Model Security / Red Team (OAuth2 auth)
	ClientID      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	TsgID         types.String `tfsdk:"tsg_id"`
	MgmtEndpoint  types.String `tfsdk:"mgmt_endpoint"`
	TokenEndpoint types.String `tfsdk:"token_endpoint"`

	// Model Security specific endpoints
	ModelSecDataEndpoint types.String `tfsdk:"model_sec_data_endpoint"`
	ModelSecMgmtEndpoint types.String `tfsdk:"model_sec_mgmt_endpoint"`

	// Red Team specific endpoints
	RedTeamDataEndpoint types.String `tfsdk:"red_team_data_endpoint"`
	RedTeamMgmtEndpoint types.String `tfsdk:"red_team_mgmt_endpoint"`
}

func (p *PrismaAIRSProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "prisma-airs"
	resp.Version = p.version
}

func (p *PrismaAIRSProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Palo Alto Networks Prisma AI Runtime Security (AIRS).",
		Attributes: map[string]schema.Attribute{
			// Scan API
			"api_key": schema.StringAttribute{
				Description: "API key for AI Runtime Security scan operations. Can also be set via PANW_AI_SEC_API_KEY.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_token": schema.StringAttribute{
				Description: "Bearer token for AI Runtime Security scan operations. Can also be set via PANW_AI_SEC_API_TOKEN.",
				Optional:    true,
				Sensitive:   true,
			},
			"profile_name": schema.StringAttribute{
				Description: "Default AI security profile name for scan operations. Can also be set via PANW_AI_SEC_PROFILE_NAME.",
				Optional:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Scan API endpoint override. Can also be set via PANW_AI_SEC_API_ENDPOINT.",
				Optional:    true,
			},

			// OAuth2
			"client_id": schema.StringAttribute{
				Description: "OAuth2 client ID for management APIs. Can also be set via PANW_MGMT_CLIENT_ID.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "OAuth2 client secret for management APIs. Can also be set via PANW_MGMT_CLIENT_SECRET.",
				Optional:    true,
				Sensitive:   true,
			},
			"tsg_id": schema.StringAttribute{
				Description: "Tenant Service Group ID. Can also be set via PANW_MGMT_TSG_ID.",
				Optional:    true,
			},
			"mgmt_endpoint": schema.StringAttribute{
				Description: "Management API endpoint override. Can also be set via PANW_MGMT_ENDPOINT.",
				Optional:    true,
			},
			"token_endpoint": schema.StringAttribute{
				Description: "OAuth2 token endpoint override. Can also be set via PANW_MGMT_TOKEN_ENDPOINT.",
				Optional:    true,
			},

			// Model Security
			"model_sec_data_endpoint": schema.StringAttribute{
				Description: "Model Security data plane endpoint. Can also be set via PANW_MODEL_SEC_DATA_ENDPOINT.",
				Optional:    true,
			},
			"model_sec_mgmt_endpoint": schema.StringAttribute{
				Description: "Model Security management plane endpoint. Can also be set via PANW_MODEL_SEC_MGMT_ENDPOINT.",
				Optional:    true,
			},

			// Red Team
			"red_team_data_endpoint": schema.StringAttribute{
				Description: "Red Team data plane endpoint. Can also be set via PANW_RED_TEAM_DATA_ENDPOINT.",
				Optional:    true,
			},
			"red_team_mgmt_endpoint": schema.StringAttribute{
				Description: "Red Team management plane endpoint. Can also be set via PANW_RED_TEAM_MGMT_ENDPOINT.",
				Optional:    true,
			},
		},
	}
}

func (p *PrismaAIRSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PrismaAIRSProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve values: provider config takes precedence over env vars
	apiKey := stringValueOrEnv(config.APIKey, "PANW_AI_SEC_API_KEY")
	apiToken := stringValueOrEnv(config.APIToken, "PANW_AI_SEC_API_TOKEN")
	profileName := stringValueOrEnv(config.ProfileName, "PANW_AI_SEC_PROFILE_NAME")
	endpoint := stringValueOrEnv(config.Endpoint, "PANW_AI_SEC_API_ENDPOINT")
	clientID := stringValueOrEnv(config.ClientID, "PANW_MGMT_CLIENT_ID")
	clientSecret := stringValueOrEnv(config.ClientSecret, "PANW_MGMT_CLIENT_SECRET")
	tsgID := stringValueOrEnv(config.TsgID, "PANW_MGMT_TSG_ID")
	mgmtEndpoint := stringValueOrEnv(config.MgmtEndpoint, "PANW_MGMT_ENDPOINT")
	tokenEndpoint := stringValueOrEnv(config.TokenEndpoint, "PANW_MGMT_TOKEN_ENDPOINT")
	modelSecDataEndpoint := stringValueOrEnv(config.ModelSecDataEndpoint, "PANW_MODEL_SEC_DATA_ENDPOINT")
	modelSecMgmtEndpoint := stringValueOrEnv(config.ModelSecMgmtEndpoint, "PANW_MODEL_SEC_MGMT_ENDPOINT")
	redTeamDataEndpoint := stringValueOrEnv(config.RedTeamDataEndpoint, "PANW_RED_TEAM_DATA_ENDPOINT")
	redTeamMgmtEndpoint := stringValueOrEnv(config.RedTeamMgmtEndpoint, "PANW_RED_TEAM_MGMT_ENDPOINT")

	providerData := &ProviderData{
		APIKey:               apiKey,
		APIToken:             apiToken,
		ProfileName:          profileName,
		Endpoint:             endpoint,
		ClientID:             clientID,
		ClientSecret:         clientSecret,
		TsgID:                tsgID,
		MgmtEndpoint:         mgmtEndpoint,
		TokenEndpoint:        tokenEndpoint,
		ModelSecDataEndpoint: modelSecDataEndpoint,
		ModelSecMgmtEndpoint: modelSecMgmtEndpoint,
		RedTeamDataEndpoint:  redTeamDataEndpoint,
		RedTeamMgmtEndpoint:  redTeamMgmtEndpoint,
	}

	// Initialize management client if OAuth2 credentials are available.
	if clientID != "" && clientSecret != "" && tsgID != "" {
		mgmtClient, err := management.NewClient(management.Opts{
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			TsgID:         tsgID,
			APIEndpoint:   mgmtEndpoint,
			TokenEndpoint: tokenEndpoint,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create management client", err.Error())
			return
		}
		providerData.MgmtClient = mgmtClient
	}

	// Initialize model security client if OAuth2 credentials are available.
	if clientID != "" && clientSecret != "" && tsgID != "" {
		msClient, err := modelsecurity.NewClient(modelsecurity.Opts{
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			TsgID:         tsgID,
			DataEndpoint:  modelSecDataEndpoint,
			MgmtEndpoint:  modelSecMgmtEndpoint,
			TokenEndpoint: tokenEndpoint,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create model security client", err.Error())
			return
		}
		providerData.ModelSecClient = msClient
	}

	// Initialize red team client if OAuth2 credentials are available.
	if clientID != "" && clientSecret != "" && tsgID != "" {
		rtClient, err := redteam.NewClient(redteam.Opts{
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			TsgID:         tsgID,
			DataEndpoint:  redTeamDataEndpoint,
			MgmtEndpoint:  redTeamMgmtEndpoint,
			TokenEndpoint: tokenEndpoint,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create red team client", err.Error())
			return
		}
		providerData.RedTeamClient = rtClient
	}

	// Initialize scan API client if API key is available.
	if apiKey != "" {
		opts := []aisec.ConfigOption{aisec.WithAPIKey(apiKey)}
		if endpoint != "" {
			opts = append(opts, aisec.WithEndpoint(endpoint))
		}
		cfg := aisec.NewConfig(opts...)
		providerData.Scanner = scan.NewScanner(cfg)
		providerData.ScanProfileName = profileName
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *PrismaAIRSProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSecurityProfileResource,
		NewCustomTopicResource,
		NewApiKeyResource,
		NewCustomerAppResource,
		NewModelSecurityGroupResource,
		NewModelScanResource,
		NewRedTeamTargetResource,
		NewRedTeamScanResource,
		NewRedTeamCustomPromptSetResource,
	}
}

func (p *PrismaAIRSProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDlpProfilesDataSource,
		NewDeploymentProfilesDataSource,
		NewScanLogsDataSource,
		NewModelSecurityRulesDataSource,
		NewModelScanEvaluationsDataSource,
		NewModelScanViolationsDataSource,
		NewRedTeamCategoriesDataSource,
		NewRedTeamQuotaDataSource,
		NewContentScanDataSource,
	}
}

// New returns a function that creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PrismaAIRSProvider{
			version: version,
		}
	}
}

// ProviderData holds resolved configuration passed to resources and data sources.
type ProviderData struct {
	MgmtClient           *management.Client
	ModelSecClient       *modelsecurity.Client
	RedTeamClient        *redteam.Client
	Scanner              *scan.Scanner
	ScanProfileName      string
	APIKey               string
	APIToken             string
	ProfileName          string
	Endpoint             string
	ClientID             string
	ClientSecret         string
	TsgID                string
	MgmtEndpoint         string
	TokenEndpoint        string
	ModelSecDataEndpoint string
	ModelSecMgmtEndpoint string
	RedTeamDataEndpoint  string
	RedTeamMgmtEndpoint  string
}

func stringValueOrEnv(val types.String, envKey string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envKey)
}
