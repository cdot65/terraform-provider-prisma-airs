package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/cdot65/prisma-airs-provider/internal/tfutil"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type field struct {
	kind, description                                          string
	required, immutable, sensitive, retained, computed, stable bool
	choices                                                    []string
}
type definition struct {
	name, description                   string
	fields                              map[string]field
	create                              func(context.Context, *client, document) (document, error)
	read                                func(context.Context, *client, string, string) (document, error)
	update                              func(context.Context, *client, string, string, document) (document, error)
	delete                              func(context.Context, *client, string, string) error
	list                                func(context.Context, *client, string, int64, int64) ([]document, int64, error)
	workspace, pagination, scopedImport bool
}

type gatewayResource struct {
	definition definition
	client     *client
}

var _ resource.ResourceWithConfigure = &gatewayResource{}
var _ resource.ResourceWithImportState = &gatewayResource{}
var _ resource.ResourceWithValidateConfig = &gatewayResource{}

func (r *gatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_" + r.definition.name
}
func (f field) attribute() schema.Attribute {
	optional, computed := !f.required && !f.computed, f.computed || (!f.required && !f.retained)
	switch f.kind {
	case "checks":
		return checksAttribute(f.description)
	case "actions":
		return actionsAttribute(f.description)
	case "bool":
		x := schema.BoolAttribute{Description: f.description, Required: f.required, Optional: optional, Computed: computed, Sensitive: f.sensitive}
		if f.immutable {
			x.PlanModifiers = []planmodifier.Bool{boolplanmodifier.RequiresReplace()}
		}
		return x
	case "number":
		x := schema.Float64Attribute{Description: f.description, Required: f.required, Optional: optional, Computed: computed, Sensitive: f.sensitive}
		if f.immutable {
			x.PlanModifiers = []planmodifier.Float64{float64planmodifier.RequiresReplace()}
		}
		return x
	case "strings":
		x := schema.SetAttribute{Description: f.description, ElementType: types.StringType, Required: f.required, Optional: optional, Computed: computed, Sensitive: f.sensitive}
		if f.immutable {
			x.PlanModifiers = []planmodifier.Set{setplanmodifier.RequiresReplace()}
		}
		return x
	case "object", "array":
		x := schema.DynamicAttribute{Description: f.description, Required: f.required, Optional: optional, Computed: computed, Sensitive: f.sensitive}
		if f.immutable {
			x.PlanModifiers = []planmodifier.Dynamic{dynamicplanmodifier.RequiresReplace()}
		}
		if f.computed && f.sensitive {
			x.PlanModifiers = []planmodifier.Dynamic{dynamicplanmodifier.UseStateForUnknown()}
		}
		return x
	default:
		x := schema.StringAttribute{Description: f.description, Required: f.required, Optional: optional, Computed: computed, Sensitive: f.sensitive}
		if f.required {
			x.Validators = []validator.String{stringvalidator.LengthAtLeast(1)}
		}
		if len(f.choices) > 0 {
			x.Validators = append(x.Validators, stringvalidator.OneOf(f.choices...))
		}
		if f.immutable {
			x.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
			if !f.required {
				x.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()}
			}
		}
		if (f.computed && (f.sensitive || f.stable)) || (f.immutable && !f.required) {
			x.PlanModifiers = append(x.PlanModifiers, stringplanmodifier.UseStateForUnknown())
		}
		return x
	}
}
func (r *gatewayResource) attributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{}
	for k, f := range r.definition.fields {
		attrs[k] = f.attribute()
	}
	return attrs
}
func (r *gatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: r.definition.description, Attributes: r.attributes()}
}
func (r *gatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, d := product.Client[*client](req.ProviderData, "gateway", "AI Gateway")
	resp.Diagnostics.Append(d...)
	r.client = c
}

func (r *gatewayResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var model types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for k, f := range r.definition.fields {
		v := model.Attributes()[k]
		if v == nil || v.IsNull() || v.IsUnknown() {
			continue
		}

		if f.kind == "object" || f.kind == "array" {
			v := v.(types.Dynamic)
			if v.IsUnderlyingValueUnknown() {
				continue
			}
			good := false
			if f.kind == "object" {
				switch v.UnderlyingValue().(type) {
				case types.Map, types.Object:
					good = true
				}
			} else {
				switch v.UnderlyingValue().(type) {
				case types.Tuple, types.List, types.Set:
					good = true
				}
			}
			if !good || v.IsUnderlyingValueNull() {
				resp.Diagnostics.AddAttributeError(path.Root(k), "Invalid native Gateway input", "Use a native HCL "+f.kind+"; JSON strings and null roots are not accepted.")
				continue
			}
			if r.definition.name == "config" && k == "config" {
				x, e := nativeJSON(v)
				if e == nil && routingCredentials(x) {
					resp.Diagnostics.AddAttributeError(path.Root(k), "Credentials in routing config", "Use integration/provider or secret-reference identifiers. Plaintext credentials are prohibited in visible routing documents.")
				}
			}
		}
	}
}

func (r *gatewayResource) body(model types.Object, update bool, diags *diag.Diagnostics) document {
	body := document{}
	for k, f := range r.definition.fields {
		if f.computed || (update && f.immutable) || k == "check_parameters" {
			continue
		}
		v := model.Attributes()[k]
		if v.IsNull() || (!f.required && !f.retained && v.IsUnknown()) {
			continue
		}
		x, e := nativeJSON(v)
		if f.kind == "checks" {
			x, e = writeChecks(v.(types.List))
		}
		if e != nil {
			diags.AddAttributeError(path.Root(k), "Unresolved Gateway input", "All configured values must be known before applying.")
			continue
		}
		body[k] = x
	}

	if configured, ok := model.Attributes()["check_parameters"]; ok && !configured.IsNull() {
		encoded, e := nativeJSON(configured)
		parameters, valid := encoded.(map[string]any)
		if e != nil || !valid {
			diags.AddAttributeError(path.Root("check_parameters"), "Invalid check parameters", "Use known native HCL parameter objects keyed by check ID.")
		} else {
			used := map[string]bool{}
			checks, _ := body["checks"].([]any)
			for _, item := range checks {
				check := item.(map[string]any)
				id, _ := check["id"].(string)
				if p, exists := parameters[id]; exists {
					if _, ok := p.(map[string]any); !ok {
						diags.AddAttributeError(path.Root("check_parameters"), "Invalid check parameters", "Each check's parameters must be a native HCL object.")
						continue
					}
					check["parameters"] = p
					used[id] = true
				}
			}
			if len(used) != len(parameters) {
				diags.AddAttributeError(path.Root("check_parameters"), "Unowned check parameters", "Every parameter object must match an ID in checks.")
			}
		}
	}
	if r.definition.name == "config" && routingCredentials(body["config"]) {
		diags.AddAttributeError(path.Root("config"), "Credentials in routing config", "Use integration/provider or secret-reference identifiers. Plaintext credentials are prohibited in visible routing documents.")
	}
	return body
}
func (r *gatewayResource) mapState(ctx context.Context, prior types.Object, remote document, receipt document, diags *diag.Diagnostics) types.Object {
	values := prior.Attributes()
	ts := prior.AttributeTypes(ctx)
	for k, f := range r.definition.fields {
		old := values[k]
		if f.retained {
			if old.IsUnknown() {
				v, e := nullNative(ctx, ts[k])
				if e == nil {
					values[k] = v
				}
			}
			continue
		}
		if f.computed && f.sensitive {
			if !old.IsNull() && !old.IsUnknown() {
				continue
			}
			if v, ok := receipt[k]; ok {
				mapped, e := readNative(ctx, v, old)
				if e == nil {
					values[k] = mapped
					continue
				}
			}
			v, e := nullNative(ctx, ts[k])
			if e == nil {
				values[k] = v
			}
			continue
		}
		value, present := remote[k]
		if !present {
			// A receipt may expose revision metadata absent from a GET model.
			value, present = receipt[k]
		}
		if !present {
			if old.IsUnknown() {
				v, e := nullNative(ctx, ts[k])
				if e == nil {
					values[k] = v
				}
			}
			continue
		}
		if r.definition.name == "config" && k == "config" && routingCredentials(value) {
			diags.AddError("Remote routing config contains credentials", "Remove embedded credentials remotely and use integration/provider or secret-reference identifiers before managing this config.")
			continue
		}
		// The deployment read contract returns 0/1 for this boolean.
		if f.kind == "bool" {
			if n, ok := value.(interface{ String() string }); ok {
				value = n.String() == "1"
			}
		}
		v, e := readNative(ctx, value, old)
		if f.kind == "checks" {
			v, e = readChecks(ctx, value, old.(types.List))
		}
		if f.kind == "actions" {
			v, e = readTypedObject(ctx, value, old.(types.Object))
		}
		if e != nil {
			diags.AddError("Cannot read Gateway state", "A Gateway response contains a value incompatible with the declared HCL type.")
			continue
		}
		values[k] = v
	}
	result, d := types.ObjectValue(ts, values)
	diags.Append(d...)
	return result
}

// Apply must honor known planned inputs even if the service adds defaults.
// Refresh remains authoritative and exposes readable remote drift afterwards.
func (r *gatewayResource) mapAppliedState(ctx context.Context, plan types.Object, remote, receipt document, diags *diag.Diagnostics) types.Object {
	result := r.mapState(ctx, plan, remote, receipt, diags)
	values := result.Attributes()
	for k, f := range r.definition.fields {
		v := plan.Attributes()[k]
		if !f.computed && !v.IsNull() && !v.IsUnknown() {
			values[k] = honorPlanned(v, values[k])
		}
	}
	result, d := types.ObjectValue(result.AttributeTypes(ctx), values)
	diags.Append(d...)
	return result
}

func gatewayError(err error) string {
	var shape *inputShapeError
	if errors.As(err, &shape) {
		return shape.Error()
	}
	var e *aisec.AISecSDKError
	if errors.As(err, &e) {
		return fmt.Sprintf("Gateway request failed (HTTP %d). Check entitlement, workspace scope and the configured settings.", e.StatusCode)
	}
	return "Gateway request or response validation failed. No response body is included because it may contain credentials."
}
func resourceIDs(model types.Object) (string, string) {
	id := model.Attributes()["id"].(types.String).ValueString()
	workspace := ""
	if v, ok := model.Attributes()["workspace_id"].(types.String); ok {
		workspace = v.ValueString()
	}
	return id, workspace
}
func archived(d document) bool {
	s, _ := d["status"].(string)
	return strings.EqualFold(s, "archived") || strings.EqualFold(s, "deleted")
}

func (r *gatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var model types.Object
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := r.body(model, false, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// SCM org writes use the numeric TSG, never the internal UUID from GET.
	body["organisation_id"] = r.client.organisation
	receipt, e := r.definition.create(ctx, r.client, body)
	if e != nil {
		resp.Diagnostics.AddError("Failed to create Gateway "+r.definition.name, gatewayError(e))
		return
	}
	id, _ := receipt["id"].(string)
	if id == "" {
		resp.Diagnostics.AddError("Missing Gateway identifier", "The create receipt omitted its resource ID; inspect the uniquely named remote object before retrying.")
		return
	}
	values := model.Attributes()
	values["id"] = types.StringValue(id)
	model = types.ObjectValueMust(model.AttributeTypes(ctx), values)
	// Persist identity and one-time secrets even if the subsequent GET fails.
	provisional := r.mapAppliedState(ctx, model, document{}, receipt, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, provisional)...)
	_, workspace := resourceIDs(model)
	remote, e := r.definition.read(ctx, r.client, id, workspace)
	if e != nil {
		resp.Diagnostics.AddError("Failed to refresh created Gateway "+r.definition.name, gatewayError(e))
		return
	}
	final := r.mapAppliedState(ctx, model, remote, receipt, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, final)...)
	}
}
func (r *gatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var model types.Object
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, w := resourceIDs(model)
	remote, e := r.definition.read(ctx, r.client, id, w)
	if tfutil.IsNotFound(e) || (e == nil && archived(remote)) {
		resp.State.RemoveResource(ctx)
		return
	}
	if e != nil {
		resp.Diagnostics.AddError("Failed to read Gateway "+r.definition.name, gatewayError(e))
		return
	}
	model = r.mapState(ctx, model, remote, nil, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	}
}
func (r *gatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var model types.Object
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, w := resourceIDs(model)
	body := r.body(model, true, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	receipt, e := r.definition.update(ctx, r.client, id, w, body)
	if e != nil {
		resp.Diagnostics.AddError("Failed to update Gateway "+r.definition.name, gatewayError(e))
		return
	}
	// An update never rewrites the resource identifier to a version identifier.
	remote, e := r.definition.read(ctx, r.client, id, w)
	if e != nil {
		resp.Diagnostics.AddError("Failed to refresh updated Gateway "+r.definition.name, gatewayError(e))
		return
	}
	model = r.mapAppliedState(ctx, model, remote, receipt, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	}
}
func (r *gatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var model types.Object
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, w := resourceIDs(model)
	e := r.definition.delete(ctx, r.client, id, w)
	if e != nil && !tfutil.IsNotFound(e) && !tfutil.IsUndecodableSuccess(e) {
		resp.Diagnostics.AddError("Failed to delete Gateway "+r.definition.name, gatewayError(e))
		return
	}
	verify, cancelVerify := context.WithTimeout(ctx, 30*time.Second)
	defer cancelVerify()
	for {
		remote, e := r.definition.read(verify, r.client, id, w)
		if tfutil.IsNotFound(e) || (e == nil && archived(remote)) {
			return
		}
		if e != nil {
			resp.Diagnostics.AddError("Failed to verify Gateway deletion", gatewayError(e))
			return
		}
		select {
		case <-verify.Done():
			resp.Diagnostics.AddError("Gateway deletion not confirmed", "The service still reports the resource as active. Check its lifecycle status before retrying destroy.")
			return
		case <-time.After(500 * time.Millisecond):
		}
	}
}
func (r *gatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, w := req.ID, ""
	if r.definition.scopedImport {
		parts := strings.Split(req.ID, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			resp.Diagnostics.AddError("Invalid scoped Gateway import", "Use workspace_id/resource_id for workspace provider or MCP server imports.")
			return
		}
		w, id = parts[0], parts[1]
	}
	remote, e := r.definition.read(ctx, r.client, id, w)
	if e != nil {
		resp.Diagnostics.AddError("Failed to import Gateway "+r.definition.name, gatewayError(e))
		return
	}
	if r.definition.scopedImport {
		if e := r.validateImportScope(ctx, remote, id, w); e != nil {
			resp.Diagnostics.AddError("Invalid Gateway import scope", gatewayError(e))
			return
		}
	}
	if archived(remote) {
		resp.Diagnostics.AddError("Cannot import archived Gateway resource", "Import an active object.")
		return
	}
	ts := map[string]attr.Type{}
	values := map[string]attr.Value{}
	for k, a := range r.attributes() {
		ts[k] = a.GetType()
		v, e := nullNative(ctx, ts[k])
		if e != nil {
			resp.Diagnostics.AddError("Cannot initialize Gateway import", e.Error())
			return
		}
		values[k] = v
	}
	values["id"] = types.StringValue(id)
	if w != "" {
		values["workspace_id"] = types.StringValue(w)
	}
	model := types.ObjectValueMust(ts, values)
	model = r.mapState(ctx, model, remote, nil, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	}
	for _, f := range r.definition.fields {
		if f.retained || (f.computed && f.sensitive) || f.kind == "checks" {
			resp.Diagnostics.AddWarning("Gateway secrets unavailable on import", "Supply desired sensitive inputs before managing them. One-time key or deployment credentials cannot be recovered by import; keep them in an external secret store or deliberately replace the resource.")
			break
		}
	}
}

// MCP detail reads can omit workspace_id. Validate membership using the scoped
// inventory rather than accepting an unverified composite import identifier.
func (r *gatewayResource) validateImportScope(ctx context.Context, remote document, id, workspace string) error {
	if actual, ok := remote["workspace_id"].(string); ok && actual != "" {
		if actual != workspace {
			return &inputShapeError{field: "workspace_id", detail: "does not match the imported resource"}
		}
		return nil
	}
	seen := map[string]bool{}
	for page := int64(1); page <= 1000; page++ {
		items, _, err := r.definition.list(ctx, r.client, workspace, 100, page)
		if err != nil {
			return err
		}
		for _, item := range items {
			itemID, _ := item["id"].(string)
			if itemID == id {
				return nil
			}
			if seen[itemID] {
				return &inputShapeError{field: "workspace_id", detail: "could not be verified because inventory repeated a page"}
			}
			seen[itemID] = true
		}
		if len(items) < 100 {
			break
		}
	}
	return &inputShapeError{field: "workspace_id", detail: "does not contain the imported resource"}
}

// Optional+computed children can be unknown inside an otherwise known object.
// Fill those from the response while retaining every known planned leaf.
func honorPlanned(plan, remote attr.Value) attr.Value {
	if plan.IsUnknown() {
		return remote
	}
	if _, err := nativeJSON(plan); err == nil {
		return plan
	}
	switch p := plan.(type) {
	case types.Object:
		r, ok := remote.(types.Object)
		if !ok || r.IsNull() || r.IsUnknown() {
			return remote
		}
		values := r.Attributes()
		for k, v := range p.Attributes() {
			values[k] = honorPlanned(v, values[k])
		}
		return types.ObjectValueMust(r.AttributeTypes(context.Background()), values)
	case types.List:
		r, ok := remote.(types.List)
		if !ok || len(p.Elements()) != len(r.Elements()) {
			return remote
		}
		values := r.Elements()
		for i, v := range p.Elements() {
			values[i] = honorPlanned(v, values[i])
		}
		return types.ListValueMust(r.ElementType(context.Background()), values)
	}
	return remote
}
