package provider

import (
	"context"
	"encoding/json"

	"github.com/cdot65/prisma-airs-go/aisec/scan"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &contentScanDataSource{}

func NewContentScanDataSource() datasource.DataSource {
	return &contentScanDataSource{}
}

type contentScanDataSource struct {
	scanner     *scan.Scanner
	profileName string
}

type ContentScanDataSourceModel struct {
	ProfileName  types.String `tfsdk:"profile_name"`
	ProfileID    types.String `tfsdk:"profile_id"`
	Prompt       types.String `tfsdk:"prompt"`
	Response     types.String `tfsdk:"response"`
	ContentCtx   types.String `tfsdk:"content_context"`
	CodePrompt   types.String `tfsdk:"code_prompt"`
	CodeResponse types.String `tfsdk:"code_response"`
	TrID         types.String `tfsdk:"tr_id"`
	SessionID    types.String `tfsdk:"session_id"`

	// Computed results
	Category         types.String `tfsdk:"category"`
	Action           types.String `tfsdk:"action"`
	ScanID           types.String `tfsdk:"scan_id"`
	ReportID         types.String `tfsdk:"report_id"`
	PromptDetected   types.String `tfsdk:"prompt_detected"`
	ResponseDetected types.String `tfsdk:"response_detected"`
}

func (d *contentScanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_scan"
}

func (d *contentScanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Performs a synchronous AI content scan.",
		Attributes: map[string]schema.Attribute{
			"profile_name": schema.StringAttribute{
				Optional:    true,
				Description: "AI security profile name. Falls back to provider profile_name.",
			},
			"profile_id": schema.StringAttribute{
				Optional:    true,
				Description: "AI security profile ID. Takes precedence over profile_name.",
			},
			"prompt": schema.StringAttribute{
				Optional:    true,
				Description: "AI prompt content to scan (max 2 MB).",
			},
			"response": schema.StringAttribute{
				Optional:    true,
				Description: "AI response content to scan (max 2 MB).",
			},
			"content_context": schema.StringAttribute{
				Optional:    true,
				Description: "Conversation context (max 100 MB).",
			},
			"code_prompt": schema.StringAttribute{
				Optional:    true,
				Description: "Code prompt to scan (max 2 MB).",
			},
			"code_response": schema.StringAttribute{
				Optional:    true,
				Description: "Code response to scan (max 2 MB).",
			},
			"tr_id": schema.StringAttribute{
				Optional:    true,
				Description: "Transaction ID (max 100 chars).",
			},
			"session_id": schema.StringAttribute{
				Optional:    true,
				Description: "Session ID (max 100 chars).",
			},
			"category": schema.StringAttribute{
				Computed:    true,
				Description: "Scan result category (benign, malicious, unknown).",
			},
			"action": schema.StringAttribute{
				Computed:    true,
				Description: "Scan result action (allow, block, alert).",
			},
			"scan_id": schema.StringAttribute{
				Computed:    true,
				Description: "Scan ID.",
			},
			"report_id": schema.StringAttribute{
				Computed:    true,
				Description: "Report ID.",
			},
			"prompt_detected": schema.StringAttribute{
				Computed:    true,
				Description: "Prompt detection results as JSON.",
			},
			"response_detected": schema.StringAttribute{
				Computed:    true,
				Description: "Response detection results as JSON.",
			},
		},
	}
}

func (d *contentScanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	scanner, profileName, diags := getScanClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.scanner = scanner
	d.profileName = profileName
}

func (d *contentScanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ContentScanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build content
	contentOpts := scan.ContentOpts{}
	if !config.Prompt.IsNull() && !config.Prompt.IsUnknown() {
		contentOpts.Prompt = config.Prompt.ValueString()
	}
	if !config.Response.IsNull() && !config.Response.IsUnknown() {
		contentOpts.Response = config.Response.ValueString()
	}
	if !config.ContentCtx.IsNull() && !config.ContentCtx.IsUnknown() {
		contentOpts.Context = config.ContentCtx.ValueString()
	}
	if !config.CodePrompt.IsNull() && !config.CodePrompt.IsUnknown() {
		contentOpts.CodePrompt = config.CodePrompt.ValueString()
	}
	if !config.CodeResponse.IsNull() && !config.CodeResponse.IsUnknown() {
		contentOpts.CodeResponse = config.CodeResponse.ValueString()
	}

	content, err := scan.NewContent(contentOpts)
	if err != nil {
		resp.Diagnostics.AddError("Invalid scan content", err.Error())
		return
	}

	// Build AI profile
	profile := scan.AiProfile{}
	if !config.ProfileID.IsNull() && !config.ProfileID.IsUnknown() {
		profile.ProfileID = config.ProfileID.ValueString()
	} else if !config.ProfileName.IsNull() && !config.ProfileName.IsUnknown() {
		profile.ProfileName = config.ProfileName.ValueString()
	} else if d.profileName != "" {
		profile.ProfileName = d.profileName
	}

	// Build scan opts
	var scanOpts []scan.SyncScanOpts
	opts := scan.SyncScanOpts{}
	hasOpts := false
	if !config.TrID.IsNull() && !config.TrID.IsUnknown() {
		opts.TrID = config.TrID.ValueString()
		hasOpts = true
	}
	if !config.SessionID.IsNull() && !config.SessionID.IsUnknown() {
		opts.SessionID = config.SessionID.ValueString()
		hasOpts = true
	}
	if hasOpts {
		scanOpts = append(scanOpts, opts)
	}

	// Perform scan
	result, err := d.scanner.SyncScan(ctx, profile, content, scanOpts...)
	if err != nil {
		resp.Diagnostics.AddError("Scan failed", err.Error())
		return
	}

	// Map results
	var state ContentScanDataSourceModel
	state.ProfileName = config.ProfileName
	state.ProfileID = config.ProfileID
	state.Prompt = config.Prompt
	state.Response = config.Response
	state.ContentCtx = config.ContentCtx
	state.CodePrompt = config.CodePrompt
	state.CodeResponse = config.CodeResponse
	state.TrID = config.TrID
	state.SessionID = config.SessionID

	state.Category = types.StringValue(result.Category)
	state.Action = types.StringValue(result.Action)
	state.ScanID = types.StringValue(result.ScanID)
	state.ReportID = types.StringValue(result.ReportID)

	if result.PromptDetected != nil {
		if b, err := json.Marshal(result.PromptDetected); err == nil {
			state.PromptDetected = types.StringValue(string(b))
		}
	} else {
		state.PromptDetected = types.StringNull()
	}

	if result.ResponseDetected != nil {
		if b, err := json.Marshal(result.ResponseDetected); err == nil {
			state.ResponseDetected = types.StringValue(string(b))
		}
	} else {
		state.ResponseDetected = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func getScanClient(data any) (*scan.Scanner, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	pd, ok := data.(*ProviderData)
	if !ok || pd == nil {
		diags.AddError("Missing provider config", "Provider not configured")
		return nil, "", diags
	}
	if pd.Scanner == nil {
		diags.AddError("Scan client not configured", "API key required (PANW_AI_SEC_API_KEY)")
		return nil, "", diags
	}
	return pd.Scanner, pd.ScanProfileName, diags
}
