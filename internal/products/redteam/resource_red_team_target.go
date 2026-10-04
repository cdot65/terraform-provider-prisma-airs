package redteam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/cdot65/prisma-airs-provider/internal/tfutil"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &redTeamTargetResource{}
	_ resource.ResourceWithImportState    = &redTeamTargetResource{}
	_ resource.ResourceWithValidateConfig = &redTeamTargetResource{}
	_ resource.ResourceWithModifyPlan     = &redTeamTargetResource{}
)

func NewRedTeamTargetResource() resource.Resource { return &redTeamTargetResource{} }

type redTeamTargetResource struct{ client *redteam.Client }
type RedTeamTargetResourceModel struct {
	ID                       types.String `tfsdk:"id"`
	UUID                     types.String `tfsdk:"uuid"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	TargetType               types.String `tfsdk:"target_type"`
	ConnectionType           types.String `tfsdk:"connection_type"`
	APIEndpointType          types.String `tfsdk:"api_endpoint_type"`
	NetworkBrokerChannelUUID types.String `tfsdk:"network_broker_channel_uuid"`
	ResponseMode             types.String `tfsdk:"response_mode"`
	OpenAI                   types.Object `tfsdk:"openai"`
	HuggingFace              types.Object `tfsdk:"hugging_face"`
	Databricks               types.Object `tfsdk:"databricks"`
	Bedrock                  types.Object `tfsdk:"bedrock"`
	Custom                   types.Object `tfsdk:"custom"`
	Rest                     types.Object `tfsdk:"rest"`
	Streaming                types.Object `tfsdk:"streaming"`
	Adapter                  types.Object `tfsdk:"adapter"`
	HeadersAuth              types.Object `tfsdk:"headers_auth"`
	BasicAuth                types.Object `tfsdk:"basic_auth"`
	OAuth2Auth               types.Object `tfsdk:"oauth2_auth"`
	UnavailableFields        types.List   `tfsdk:"unavailable_fields"`
	Status                   types.String `tfsdk:"status"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
}

func (m *RedTeamTargetResourceModel) blocks() map[string]*types.Object {
	return map[string]*types.Object{"openai": &m.OpenAI, "hugging_face": &m.HuggingFace, "databricks": &m.Databricks, "bedrock": &m.Bedrock, "custom": &m.Custom, "rest": &m.Rest, "streaming": &m.Streaming, "adapter": &m.Adapter, "headers_auth": &m.HeadersAuth, "basic_auth": &m.BasicAuth, "oauth2_auth": &m.OAuth2Auth}
}
func (r *redTeamTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_target"
}
func (r *redTeamTargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages a Red Team target with exactly one native connection block. Desired payloads and secrets are preserved through masked reads; arbitrary payload drift cannot be detected reliably.", Attributes: map[string]schema.Attribute{
		"id":                          schema.StringAttribute{Computed: true, Description: "Terraform resource ID (same as uuid).", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"uuid":                        schema.StringAttribute{Computed: true, Description: "Target UUID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":                        schema.StringAttribute{Required: true, Description: "Target name.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"description":                 schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Target description."},
		"target_type":                 schema.StringAttribute{Optional: true, Computed: true, Description: "Target category.", Validators: []validator.String{stringvalidator.OneOf("APPLICATION", "AGENT", "MODEL")}},
		"connection_type":             schema.StringAttribute{Computed: true, Description: "Provider inferred from the selected connection block."},
		"response_mode":               schema.StringAttribute{Computed: true, Description: "REST or STREAMING for endpoint connections; null for adapter-controlled transport."},
		"api_endpoint_type":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("PUBLIC"), Description: "Endpoint accessibility.", Validators: []validator.String{stringvalidator.OneOf("PUBLIC", "PRIVATE", "NETWORK_BROKER")}},
		"network_broker_channel_uuid": targetString(false, "Preexisting Network Broker channel UUID. This resource never creates channels."),
		"unavailable_fields":          schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Payload/header paths explicitly redacted by the API. Writes require original values at these paths; no-op adoption does not."},
		"status":                      schema.StringAttribute{Computed: true, Description: "Target status."},
		"created_at":                  schema.StringAttribute{Computed: true, Description: "Creation timestamp.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"updated_at":                  schema.StringAttribute{Computed: true, Description: "Update timestamp."},
	}, Blocks: targetBlocks()}
}

// The live update API ignores null authentication, retaining the old secret.
// Removing authentication must therefore recreate the target to actually clear it.
func (r *redTeamTargetResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var state, plan RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	priorFamily, priorBlock, _ := selectedTargetBlock(&state, targetFamilies)
	family, block, unknownFamily := selectedTargetBlock(&plan, targetFamilies)
	if !unknownFamily && family != "" && family != "multiple" {
		connection := strings.ToUpper(family)
		if family == "adapter" {
			connection = "CUSTOM_TARGET_ADAPTER"
		}
		mode := "REST"
		targetType := "MODEL"
		if family == "adapter" {
			targetType = "APPLICATION"
		}
		if endpointTargetFamily(family) {
			connection = "CUSTOM"
			targetType = "APPLICATION"
		}
		if family == "streaming" || family == "databricks" {
			mode = "STREAMING"
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("connection_type"), types.StringValue(connection))...)
		plannedMode := types.StringValue(mode)
		if family == "adapter" {
			plannedMode = types.StringNull()
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("response_mode"), plannedMode)...)
		if plan.TargetType.IsUnknown() {
			plan.TargetType = types.StringValue(targetType)
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("target_type"), plan.TargetType)...)
		}
		priorNative := nativeTargetFamily(priorFamily)
		native := nativeTargetFamily(family)
		if priorFamily != family && (priorNative || native || priorFamily == "adapter" || family == "adapter") {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root(family))
		}
		if priorFamily == family && !priorBlock.IsNull() && !priorBlock.IsUnknown() && !block.IsNull() && !block.IsUnknown() {
			// Masked native credentials cannot prove an update took effect. Recreate
			// for any native connection change; ordinary metadata edits still update.
			if native && !priorBlock.Equal(block) {
				resp.RequiresReplace = append(resp.RequiresReplace, path.Root(family))
			}
			for field, before := range priorBlock.Attributes() {
				after := block.Attributes()[field]
				if !before.IsNull() && !after.IsUnknown() && after.IsNull() {
					resp.RequiresReplace = append(resp.RequiresReplace, path.Root(family).AtName(field))
				}
			}
		}
	}
	if !plan.TargetType.IsNull() && !plan.TargetType.IsUnknown() && !plan.TargetType.Equal(state.TargetType) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("target_type"))
	}
	if !state.NetworkBrokerChannelUUID.IsNull() && !plan.NetworkBrokerChannelUUID.IsUnknown() && plan.NetworkBrokerChannelUUID.IsNull() {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("network_broker_channel_uuid"))
	}
	for _, name := range targetAuthBlocks {
		before, after := *state.blocks()[name], *plan.blocks()[name]
		if before.IsNull() || before.IsUnknown() || after.IsNull() || after.IsUnknown() {
			continue
		}
		for field, value := range before.Attributes() {
			if !value.IsNull() && !after.Attributes()[field].IsUnknown() && after.Attributes()[field].IsNull() {
				resp.RequiresReplace = append(resp.RequiresReplace, path.Root(name).AtName(field))
			}
		}
	}
	prior, _, _ := selectedTargetBlock(&state, targetAuthBlocks)
	desired, _, unknown := selectedTargetBlock(&plan, targetAuthBlocks)
	if prior != "" && desired == "" && !unknown {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root(prior))
	}
}
func (r *redTeamTargetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !resp.Diagnostics.HasError() {
		validateTargetConfig(&config, &resp.Diagnostics)
	}
}
func (r *redTeamTargetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getRedTeamClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}
func (r *redTeamTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := targetRequest(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	target, err := r.client.Targets.CreateDetails(ctx, body, false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create red team target", safeTargetError(err))
		return
	}
	recordTargetReceipt(&plan, target.UUID, string(target.Status), target.CreatedAt, target.UpdatedAt, body)
	r.readDesiredTarget(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *redTeamTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	target, err := r.client.Targets.GetDetails(ctx, state.UUID.ValueString())
	if tfutil.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to read red team target", safeTargetError(err))
		return
	}
	mapTargetDetailsToState(ctx, target, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *redTeamTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateTargetUnavailable(&state, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	create := targetRequest(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// The generated create/update contracts are structurally identical. Conversion
	// through their JSON methods preserves explicit nulls used to clear old auth.
	encoded, err := json.Marshal(create)
	if err != nil {
		resp.Diagnostics.AddError("Invalid target configuration", "Unable to serialize target settings.")
		return
	}
	var update rtschema.TargetUpdateRequest
	if err = json.Unmarshal(encoded, &update); err != nil {
		resp.Diagnostics.AddError("Invalid target configuration", "Unable to encode typed update settings.")
		return
	}
	target, err := r.client.Targets.UpdateDetails(ctx, state.UUID.ValueString(), update, false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update red team target", safeTargetError(err))
		return
	}
	plan.CreatedAt = state.CreatedAt
	recordTargetReceipt(&plan, state.UUID.ValueString(), string(target.Status), target.CreatedAt, target.UpdatedAt, create)
	r.readDesiredTarget(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *redTeamTargetResource) readDesiredTarget(ctx context.Context, state *RedTeamTargetResourceModel, diags *diag.Diagnostics) {
	target, err := r.client.Targets.GetDetails(ctx, state.UUID.ValueString())
	if err != nil {
		diags.AddError("Failed to read target after write", safeTargetError(err))
		return
	}
	mapTargetDetailsToState(ctx, target, state, diags)
}
func (r *redTeamTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedTeamTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.Targets.Delete(ctx, state.UUID.ValueString())
	tfutil.FinishDelete(ctx, err, "red team target", func(ctx context.Context) (bool, error) {
		_, err := r.client.Targets.GetDetails(ctx, state.UUID.ValueString())
		if tfutil.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}, &resp.Diagnostics)
}
func (r *redTeamTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	familyHint := ""
	if parts := strings.SplitN(id, "/", 2); len(parts) == 2 {
		familyHint, id = parts[0], parts[1]
		valid := false
		for _, family := range targetFamilies {
			if family == familyHint {
				valid = true
			}
		}
		if !valid {
			resp.Diagnostics.AddError("Invalid target import hint", "Use <uuid> or <connection_block>/<uuid>.")
			return
		}
	}
	target, err := r.client.Targets.GetDetails(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import target", safeTargetError(err))
		return
	}
	var state RedTeamTargetResourceModel
	for name, block := range state.blocks() {
		*block = types.ObjectNull(targetObjectTypes(name))
	}
	if familyHint != "" {
		*state.blocks()[familyHint] = types.ObjectValueMust(targetObjectTypes(familyHint), emptyTargetAttributes(familyHint))
	}
	mapTargetDetailsToState(ctx, target, &state, &resp.Diagnostics)
	if familyHint != "" {
		actual, _, _ := selectedTargetBlock(&state, targetFamilies)
		if actual != familyHint {
			resp.Diagnostics.AddError("Target import family mismatch", "The requested connection block does not match this target's provider and response mode.")
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.AddWarning("Imported target secrets unavailable", "Observable payloads are recovered. Missing or masked credentials remain null and can be kept for a read-only no-op plan. Supply complete original inputs before any create or update.")
}
func safeTargetError(err error) string {
	// Target service errors can echo request credentials. Return only typed error
	// metadata, never the request/response body or a raw wrapped error string.
	var sdk *aisec.AISecSDKError
	if errors.As(err, &sdk) {
		return fmt.Sprintf("AIRS target request failed (HTTP %d, %s). Check endpoint, provider, authentication and tenant permissions.", sdk.StatusCode, sdk.ErrorType.String())
	}
	return "AIRS target request failed. Check endpoint, provider, authentication and tenant permissions."
}
func recordTargetReceipt(state *RedTeamTargetResourceModel, id, status, created, updated string, body rtschema.TargetCreateRequest) {
	state.ID = types.StringValue(id)
	state.UUID = state.ID
	if created != "" || state.CreatedAt.IsNull() || state.CreatedAt.IsUnknown() {
		state.CreatedAt = types.StringValue(created)
	}
	state.UpdatedAt = types.StringValue(updated)
	state.Status = types.StringValue(status)
	if value, ok := body.ConnectionType.Get(); ok {
		state.ConnectionType = types.StringValue(string(value))
	}
	if value, ok := body.TargetType.Get(); ok {
		state.TargetType = types.StringValue(string(value))
	}
	if value, ok := body.ResponseMode.Get(); ok {
		state.ResponseMode = types.StringValue(string(value))
	}
}
