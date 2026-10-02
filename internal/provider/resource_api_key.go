package provider

import (
	"context"
	"fmt"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

func NewApiKeyResource() resource.Resource {
	return &apiKeyResource{}
}

type apiKeyResource struct {
	client *airsruntime.Client
}

type ApiKeyResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ApiKeyID             types.String `tfsdk:"api_key_id"`
	ApiKeyName           types.String `tfsdk:"api_key_name"`
	ApiKey               types.String `tfsdk:"api_key"`
	AuthCode             types.String `tfsdk:"auth_code"`
	CustEnv              types.String `tfsdk:"cust_env"`
	CustCloudProvider    types.String `tfsdk:"cust_cloud_provider"`
	CustAIAgentFramework types.String `tfsdk:"cust_ai_agent_framework"`
	CustApp              types.String `tfsdk:"cust_app"`
	CreatedBy            types.String `tfsdk:"created_by"`
	Status               types.String `tfsdk:"status"`
	Revoked              types.Bool   `tfsdk:"revoked"`
	CreatedAt            types.String `tfsdk:"created_at"`
	ExpiresAt            types.String `tfsdk:"expires_at"`
	RotationTimeInterval types.Int64  `tfsdk:"rotation_time_interval"`
	RotationTimeUnit     types.String `tfsdk:"rotation_time_unit"`
}

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AI Runtime Security API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform resource ID (same as api_key_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_key_name": schema.StringAttribute{
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 31)},
				Description: "Name of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"api_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The API key value. Only available at creation time.",
			},
			"auth_code": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Deployment profile auth code for API key creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cust_app": schema.StringAttribute{
				Optional: true, Computed: true,
				Description:   "Customer application name to associate with the key.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"created_by": schema.StringAttribute{
				Optional: true, Computed: true,
				Description:   "Identity of the user creating the key.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"rotation_time_interval": schema.Int64Attribute{
				Required:      true,
				Description:   "Rotation interval value (e.g. 90 for 90 days).",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{int64validator.Between(1, 2147483647)},
			},
			"rotation_time_unit": schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{stringvalidator.OneOf("days", "months")},
				Description:   "Rotation time unit (days, months).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cust_env": schema.StringAttribute{
				Optional: true, Computed: true, Description: "Customer environment used when creating the associated application.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"cust_cloud_provider": schema.StringAttribute{
				Optional: true, Computed: true, Description: "Customer cloud provider used when creating the associated application.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"cust_ai_agent_framework": schema.StringAttribute{
				Optional: true, Computed: true, Description: "Customer AI agent framework.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "API key status.",
			},
			"revoked": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the API key is revoked.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "Expiration timestamp.",
			},
		},
	}
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getMgmtClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = client
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	custApp := ""
	if !plan.CustApp.IsNull() && !plan.CustApp.IsUnknown() {
		custApp = plan.CustApp.ValueString()
	}

	createReq := airsruntime.CreateApiKeyRequest{
		ApiKeyName:           plan.ApiKeyName.ValueString(),
		AuthCode:             plan.AuthCode.ValueString(),
		CustApp:              custApp,
		CustEnv:              plan.CustEnv.ValueString(),
		CustCloudProvider:    plan.CustCloudProvider.ValueString(),
		CustAIAgentFramework: plan.CustAIAgentFramework.ValueString(),
		CreatedBy:            plan.CreatedBy.ValueString(),
		RotationTimeInterval: int32(plan.RotationTimeInterval.ValueInt64()),
		RotationTimeUnit:     plan.RotationTimeUnit.ValueString(),
	}

	key, err := r.client.ApiKeys.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	mapApiKeyToState(key, &plan)
	// Preserve the create receipt and one-time secret even if canonical reread fails.
	for _, value := range []*types.String{&plan.CustApp, &plan.CustEnv, &plan.CustCloudProvider, &plan.CustAIAgentFramework, &plan.CreatedBy} {
		if value.IsUnknown() {
			*value = types.StringNull()
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	found, err := findApiKeyByID(ctx, r.client, key.ApiKeyID)
	if err != nil || found == nil {
		resp.Diagnostics.AddError("Failed to verify created API key", "The creation receipt was saved, but canonical key metadata could not be read. Refresh before retrying.")
		return
	}
	mapApiKeyMetadata(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := findApiKeyByID(ctx, r.client, state.ApiKeyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read API key", err.Error())
		return
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Only the one-time secret is unavailable after create. Readable creation
	// metadata follows the canonical list response; an imported secret stays null.
	secret := state.ApiKey
	mapApiKeyToState(found, &state)
	state.ApiKey = secret
	mapApiKeyMetadata(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// api_key_name has ForceNew, so Update should never be called.
	resp.Diagnostics.AddError("Update not supported", "API keys cannot be updated in place; they must be recreated.")
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdBy := state.CreatedBy.ValueString()
	if createdBy == "" {
		createdBy = "terraform"
	}

	_, err := r.client.ApiKeys.Delete(ctx, state.ApiKeyName.ValueString(), createdBy)
	finishDelete(ctx, err, "API key", func(ctx context.Context) (bool, error) {
		found, lookupErr := findApiKeyByID(ctx, r.client, state.ApiKeyID.ValueString())
		return found == nil, lookupErr
	}, &resp.Diagnostics)
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	apiKeyID := req.ID

	found, err := findApiKeyByID(ctx, r.client, apiKeyID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import API key", err.Error())
		return
	}
	if found == nil {
		resp.Diagnostics.AddError("API key not found", "No API key with ID: "+apiKeyID)
		return
	}

	var state ApiKeyResourceModel
	mapApiKeyToState(found, &state)
	// The creation-time api_key is unavailable during import.
	state.ApiKey = types.StringNull()
	mapApiKeyMetadata(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// findApiKeyByID searches for an API key by ID using paginated list calls. A
// nil key with a nil error means the key does not exist; list failures are
// returned so a transient error is never mistaken for a deleted key.
func findApiKeyByID(ctx context.Context, client *airsruntime.Client, apiKeyID string) (*airsruntime.ApiKey, error) {
	offset := 0
	limit := 100
	seen := map[string]bool{}
	for {
		listResp, err := client.ApiKeys.List(ctx, airsruntime.ListOpts{Limit: limit, Offset: offset})
		if err != nil {
			return nil, err
		}
		for i := range listResp.Items {
			if listResp.Items[i].ApiKeyID == apiKeyID {
				return &listResp.Items[i], nil
			}
		}
		if len(listResp.Items) == 0 && listResp.NextOffset > offset {
			return nil, fmt.Errorf("key list returned an empty advancing page")
		}
		for _, key := range listResp.Items {
			if key.ApiKeyID == "" || seen[key.ApiKeyID] {
				return nil, fmt.Errorf("key list omitted identity or repeated a page")
			}
			seen[key.ApiKeyID] = true
		}
		if listResp.NextOffset > offset {
			offset = listResp.NextOffset
		} else if len(listResp.Items) >= limit {
			offset += len(listResp.Items)
		} else {
			return nil, nil
		}
	}
}

func mapApiKeyToState(key *airsruntime.ApiKey, state *ApiKeyResourceModel) {
	state.ID = types.StringValue(key.ApiKeyID)
	state.ApiKeyID = types.StringValue(key.ApiKeyID)
	state.ApiKeyName = types.StringValue(key.ApiKeyName)
	state.Revoked = types.BoolValue(key.Revoked)
	state.Status = types.StringValue(key.Status)
	state.CreatedAt = types.StringValue(key.CreationTS)
	state.ExpiresAt = types.StringValue(key.Expiration)
	if key.ApiKey != "" {
		state.ApiKey = types.StringValue(key.ApiKey)
	}
}

func mapApiKeyMetadata(key *airsruntime.ApiKey, state *ApiKeyResourceModel) {
	state.AuthCode = types.StringValue(key.AuthCode)
	state.CustApp = optionalString(key.CustApp)
	state.CustEnv = optionalString(key.CustEnv)
	state.CustCloudProvider = optionalString(key.CustCloudProvider)
	state.CustAIAgentFramework = optionalString(key.CustAIAgentFramework)
	state.CreatedBy = optionalString(key.CreatedBy)
	state.RotationTimeInterval = types.Int64Value(int64(key.RotationTimeInterval))
	state.RotationTimeUnit = types.StringValue(key.RotationTimeUnit)
}

func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
