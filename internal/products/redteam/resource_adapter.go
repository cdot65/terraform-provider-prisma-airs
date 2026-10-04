package redteam

import (
	"context"

	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/cdot65/prisma-airs-provider/internal/tfutil"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type adapterResource struct{ client *redteam.Client }

var _ resource.ResourceWithImportState = &adapterResource{}
var _ resource.ResourceWithConfigure = &adapterResource{}

func NewAdapterResource() resource.Resource { return &adapterResource{} }
func (r *adapterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_adapter"
}
func (r *adapterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages a Red Team adapter script and its complete variable key set. Defaults to draft writes without execution. Explicit validate=true executes through an existing Network Broker channel.", Attributes: map[string]schema.Attribute{
		"id":                          schema.StringAttribute{Computed: true, Description: "Adapter UUID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":                        schema.StringAttribute{Required: true, Description: "Adapter name.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"description":                 schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: "Adapter description."},
		"script":                      schema.StringAttribute{Required: true, Sensitive: true, Description: "Plaintext Python adapter script. Base64 encoding is internal; scripts may contain credentials.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"network_broker_channel_uuid": targetString(false, "Existing Network Broker channel; required to execute validation. Terraform does not install or manage the broker."),
		"variables": schema.MapNestedAttribute{Optional: true, Computed: true, Sensitive: true, Default: mapdefault.StaticValue(types.MapValueMust(types.ObjectType{AttrTypes: adapterVariableTypes}, map[string]attr.Value{})), Description: "Complete desired key set. Omitted keys are deleted on update. Null preserves an existing SECRET with the same key/type.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{Required: true, Description: "VAR or SECRET.", Validators: []validator.String{stringvalidator.OneOf("VAR", "SECRET")}},
			"value": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Desired value, or null to retain an imported SECRET. New variables need values."},
		}}},
		"validate":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Explicit execution opt-in on create/update. False saves DRAFT; true validates through Network Broker. Import sets true only for ACTIVE adapters without executing them."},
		"validation_prompt": schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString("Terraform adapter validation"), Description: "Transient prompt sent when validate=true; not recovered on import."},
		"status":            schema.StringAttribute{Computed: true, Description: "DRAFT or ACTIVE."},
		"created_at":        schema.StringAttribute{Computed: true, Description: "Creation timestamp.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"updated_at":        schema.StringAttribute{Computed: true, Description: "Update timestamp."},
	}}
}
func (r *adapterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		c, d := getRedTeamClient(req.ProviderData)
		resp.Diagnostics.Append(d...)
		r.client = c
	}
}
func (r *adapterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan adapterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := adapterRequest(&plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	execute := plan.Validate.ValueBool()
	remote, err := r.client.Adapters.Create(ctx, body, redteam.AdapterWriteOpts{Validate: &execute})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create adapter", safeTargetError(err))
		return
	}
	mapAdapter(ctx, remote, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *adapterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state adapterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Adapters.Get(ctx, state.ID.ValueString())
	if tfutil.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to read adapter", safeTargetError(err))
		return
	}
	mapAdapter(ctx, remote, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *adapterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state adapterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := adapterRequest(&plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	update := rtschema.CustomTargetAdapterUpdateRequest(body)
	execute := plan.Validate.ValueBool()
	remote, err := r.client.Adapters.Update(ctx, state.ID.ValueString(), update, redteam.AdapterWriteOpts{Validate: &execute})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update adapter", safeTargetError(err))
		return
	}
	mapAdapter(ctx, remote, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *adapterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state adapterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Adapters.Delete(ctx, state.ID.ValueString())
	tfutil.FinishDelete(ctx, err, "red team adapter", func(ctx context.Context) (bool, error) {
		_, err := r.client.Adapters.Get(ctx, state.ID.ValueString())
		if tfutil.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}, &resp.Diagnostics)
}
func (r *adapterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	remote, err := r.client.Adapters.Get(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import adapter", safeTargetError(err))
		return
	}
	state := adapterModel{Validate: types.BoolValue(remote.Status == rtschema.CustomTargetAdapterStatusActive), ValidationPrompt: types.StringValue("Terraform adapter validation")}
	mapAdapter(ctx, remote, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.AddWarning("Imported adapter execution setting", "Import and refresh do not execute scripts. Match validate to the imported state for a no-op plan; future writes with validate=true execute the script. Keep every variable key, using null for redacted secrets.")
}
