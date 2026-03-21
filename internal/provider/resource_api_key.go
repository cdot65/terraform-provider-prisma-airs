package provider

import (
	"context"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	client *management.Client
}

type ApiKeyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ApiKeyID   types.String `tfsdk:"api_key_id"`
	ApiKeyName types.String `tfsdk:"api_key_name"`
	ApiKey     types.String `tfsdk:"api_key"`
	CreatedBy  types.String `tfsdk:"created_by"`
	Status     types.String `tfsdk:"status"`
	Revoked    types.Bool   `tfsdk:"revoked"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
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
			"created_by": schema.StringAttribute{
				Optional:    true,
				Description: "Identity of the user creating the key.",
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

	createReq := management.CreateApiKeyRequest{
		ApiKeyName: plan.ApiKeyName.ValueString(),
		CreatedBy:  plan.CreatedBy.ValueString(),
	}

	key, err := r.client.ApiKeys.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	mapApiKeyToState(key, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := findApiKeyByID(ctx, r.client, state.ApiKeyID.ValueString())
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve write-only values from state since they're not returned by the API.
	apiKeyVal := state.ApiKey
	createdByVal := state.CreatedBy
	mapApiKeyToState(found, &state)
	state.ApiKey = apiKeyVal
	state.CreatedBy = createdByVal
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
	if err != nil {
		if strings.Contains(err.Error(), "failed to parse response JSON") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete API key", err.Error())
		return
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	apiKeyID := req.ID

	found := findApiKeyByID(ctx, r.client, apiKeyID)
	if found == nil {
		resp.Diagnostics.AddError("API key not found", "No API key with ID: "+apiKeyID)
		return
	}

	var state ApiKeyResourceModel
	mapApiKeyToState(found, &state)
	// api_key value is not available during import
	state.ApiKey = types.StringValue("")
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// findApiKeyByID searches for an API key by ID using paginated list calls.
func findApiKeyByID(ctx context.Context, client *management.Client, apiKeyID string) *management.ApiKey {
	offset := 0
	limit := 100
	for {
		listResp, err := client.ApiKeys.List(ctx, management.ListOpts{Limit: limit, Offset: offset})
		if err != nil {
			return nil
		}
		for i := range listResp.Items {
			if listResp.Items[i].ApiKeyID == apiKeyID {
				return &listResp.Items[i]
			}
		}
		if len(listResp.Items) < limit {
			return nil
		}
		offset += limit
	}
}

func mapApiKeyToState(key *management.ApiKey, state *ApiKeyResourceModel) {
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
