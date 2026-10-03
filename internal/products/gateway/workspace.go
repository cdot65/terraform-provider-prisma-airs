package gateway

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cdot65/prisma-airs-go/aisec"
	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	parity "github.com/cdot65/prisma-airs-go/aisec/parity/schema"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// workspaceResource owns the dedicated-scope workflow. Other Gateway resources
// consume its workspace UUID without knowing the provisioning steps.
type workspaceResource struct{ client *client }

var _ resource.ResourceWithConfigure = &workspaceResource{}

func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_workspace"
}
func (r *workspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"name":                  schema.StringAttribute{Required: true, Description: "Workspace display label. Renaming preserves its UUID and unique slug.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"scope_name":            schema.StringAttribute{Required: true, Description: "Stable IAM scope name, rather than its composite display ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.RegexMatches(workspaceScopePattern, "Use a scope name, never a composite scope ID.")}},
		"scope_management":      schema.StringAttribute{Required: true, Description: "managed creates and cleans a dedicated scope; external never writes IAM. External owners maintain bindings and role grants.", Validators: []validator.String{stringvalidator.OneOf("managed", "external")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"description":           schema.StringAttribute{Optional: true, Description: "Workspace description. The API does not clear blank descriptions; removal stops managing this field.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"icon":                  schema.StringAttribute{Optional: true, Description: "Workspace icon. Removing a previously configured value clears it."},
		"scope_ownership_token": schema.StringAttribute{Computed: true, Description: "Correlation token saved before scope creation; never a credential or permission grant."},
		"scope_owned":           schema.BoolAttribute{Computed: true, Description: "Whether Terraform acknowledged ownership of the dedicated IAM scope."},
		"scope_binding_ready":   schema.BoolAttribute{Computed: true, Description: "Whether the scope is bound to this workspace slug. This does not grant service-account access."},
		"provisioning_stage":    schema.StringAttribute{Computed: true, Description: "Last completed provisioning or cleanup stage; retained after partial failures."},
	}
	for k, a := range workspaceSettingsAttributes() {
		attrs[k] = a
	}
	for k, desc := range map[string]string{"id": "Workspace UUID; a pending recovery identity may be present after a failed create.", "slug": "Server-assigned unique workspace slug.", "status": "Remote workspace lifecycle status.", "created_at": "Creation timestamp.", "last_updated_at": "Last update timestamp."} {
		attribute := schema.StringAttribute{Computed: true, Description: desc}
		if k == "id" || k == "slug" {
			attribute.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		}
		attrs[k] = attribute
	}
	resp.Schema = schema.Schema{Description: "Coordinates Gateway workspace creation and archival with dedicated IAM scope ownership. External scopes receive no writes. Membership and role assignments are separately managed.", Attributes: attrs}
}
func (r *workspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, d := product.Client[*client](req.ProviderData, "gateway", "AI Gateway")
	resp.Diagnostics.Append(d...)
	r.client = c
}
func workspaceValue(m types.Object, k string) string {
	v, _ := m.Attributes()[k].(types.String)
	return v.ValueString()
}
func workspaceFlag(m types.Object, k string) bool {
	v, _ := m.Attributes()[k].(types.Bool)
	return v.ValueBool()
}
func workspaceChange(m types.Object, k string, v any) types.Object {
	a := m.Attributes()
	switch x := v.(type) {
	case string:
		a[k] = types.StringValue(x)
	case bool:
		a[k] = types.BoolValue(x)
	}
	return types.ObjectValueMust(m.AttributeTypes(context.Background()), a)
}
func workspaceInitial(m types.Object) types.Object {
	a := m.Attributes()
	for k, v := range a {
		if v.IsUnknown() {
			n, e := nullNative(context.Background(), v.Type(context.Background()))
			if e == nil {
				a[k] = n
			}
		}
	}
	a["scope_owned"] = types.BoolValue(false)
	a["scope_binding_ready"] = types.BoolValue(false)
	a["provisioning_stage"] = types.StringValue("initial")
	return types.ObjectValueMust(m.AttributeTypes(context.Background()), a)
}
func workspacePayload(m types.Object, create bool) (document, error) {
	b := document{}
	keys := []string{"name", "description", "icon", "defaults", "usage_limits", "rate_limits"}
	if create {
		keys = append(keys, "scope_name")
	}
	for _, k := range keys {
		v := m.Attributes()[k]
		if v.IsNull() {
			continue
		}
		x, e := workspaceInput(v)
		if e != nil {
			return nil, e
		}
		if create && k == "usage_limits" && len(x.([]any)) == 0 {
			continue
		}
		b[k] = x
	}
	return b, nil
}
func workspaceDecode[Q any](b document) (Q, error) {
	var q Q
	raw, e := json.Marshal(b)
	if e == nil {
		e = json.Unmarshal(raw, &q)
	}
	return q, e
}
func (r *workspaceResource) detail(ctx context.Context, id string) (document, error) {
	return readSDK(r.client.sdk.Workspaces.Get(ctx, id, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin}))
}
func workspaceState(m types.Object, remote document, applied bool) types.Object {
	a := m.Attributes()
	for _, k := range []string{"id", "slug", "created_at", "last_updated_at"} {
		if v, ok := remote[k].(string); ok {
			a[k] = types.StringValue(v)
		}
	}
	if !applied {
		if label := workspaceLabel(remote); label != "" {
			a["name"] = types.StringValue(label)
		}
	}
	status, _ := remote["status"].(string)
	if status == "" {
		status = "active"
	}
	a["status"] = types.StringValue(status)

	return types.ObjectValueMust(m.AttributeTypes(context.Background()), a)
}
func (r *workspaceResource) fail(d *diag.Diagnostics, action string, e error) {
	d.AddError("Gateway workspace "+action+" failed", gatewayError(e)+" Completed identities remain in state when available. Inspect provisioning_stage before retrying; a failed create may require terraform untaint to resume rather than replacement.")
}

var workspaceScopePattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`)

func workspaceCheckpoint(ctx context.Context, state *tfsdk.State, m types.Object, d *diag.Diagnostics) bool {
	d.Append(state.Set(ctx, m)...)
	return !d.HasError()
}
func workspaceScopeDescription(m types.Object) string {
	return "Terraform workspace ownership " + workspaceValue(m, "scope_ownership_token")
}
func workspaceReady(m types.Object) types.Object {
	return workspaceChange(m, "provisioning_stage", "ready")
}
func (r *workspaceResource) validateCreate(m types.Object) (gw.WorkspaceCreateRequest, error) {
	var empty gw.WorkspaceCreateRequest
	if !workspaceScopePattern.MatchString(workspaceValue(m, "scope_name")) {
		return empty, &inputShapeError{field: "scope_name", detail: "must be a scope name, never a composite ID"}
	}
	b, err := workspacePayload(m, true)
	if err != nil {
		return empty, err
	}
	q, err := workspaceDecode[gw.WorkspaceCreateRequest](b)
	if err != nil {
		return empty, &inputShapeError{field: "settings", detail: "do not match the workspace contract"}
	}
	if err = parity.Validate("GatewayWorkspaceCreateRequestSchema", q); err != nil {
		return empty, &inputShapeError{field: "settings", detail: "do not satisfy the workspace contract"}
	}
	return q, nil
}

// recoverScope only acknowledges an uncertain POST when the unique intent marker
// and dedicated identity match. It never adopts an unrelated name collision.
func (r *workspaceResource) recoverScope(ctx context.Context, m types.Object, d *diag.Diagnostics) (types.Object, bool) {
	scope, err := r.client.sdk.IAMScopes.Get(ctx, workspaceValue(m, "scope_name"))
	if aisec.IsNotFound(err) {
		return m, false
	}
	if err != nil {
		r.fail(d, "uncertain scope lookup", err)
		return m, false
	}
	if workspaceValue(m, "scope_ownership_token") == "" || scope.Description != workspaceScopeDescription(m) {
		d.AddError("Scope creation outcome cannot be owned", "The scope does not match the retained creation token. No IAM writes are permitted; resolve the name collision explicitly.")
		return m, false
	}
	if !dedicatedScope(scope, workspaceValue(m, "scope_name"), "", r.client.organisation, false, d) {
		return m, false
	}
	m = workspaceChange(m, "scope_owned", true)
	return workspaceChange(m, "provisioning_stage", "scope_ready"), true
}
func (r *workspaceResource) reconcileWorkspace(ctx context.Context, m types.Object, d *diag.Diagnostics) (types.Object, bool) {
	found, err := r.findScopeWorkspace(ctx, workspaceValue(m, "scope_name"))
	if err != nil {
		r.fail(d, "partial-create reconciliation", err)
		return m, false
	}
	if found == nil {
		d.AddError("Workspace creation outcome is unresolved", "A complete inventory has no unique workspace for the recorded scope. The POST is not replayed and the scope is retained. Recover the exact UUID with import after inspecting the creation outcome.")
		return m, false
	}
	m = workspaceChange(m, "id", found["id"].(string))
	m = workspaceChange(m, "slug", found["slug"].(string))
	m = workspaceChange(m, "provisioning_stage", "workspace_created")
	if found["status"] == "archived" {
		m = workspaceChange(m, "status", "archived")
		m = workspaceChange(m, "provisioning_stage", "archived")
	}
	return m, true
}
func (r *workspaceResource) bindScope(ctx context.Context, m types.Object, d *diag.Diagnostics) (types.Object, bool) {
	name, slug := workspaceValue(m, "scope_name"), workspaceValue(m, "slug")
	scope, err := r.client.sdk.IAMScopes.Get(ctx, name)
	if err != nil {
		r.fail(d, "scope binding preflight", err)
		return m, false
	}
	if workspaceFlag(m, "scope_owned") {
		if !dedicatedScope(scope, name, slug, r.client.organisation, false, d) {
			return m, false
		}
		if len(scope.Resources) == 0 {
			_, err = r.client.sdk.IAMScopes.Update(ctx, name, gw.IAMScopeUpdateInput{Description: scope.Description, Resources: []gw.IAMScopeResource{{ResourceType: "workspace", ResourceID: slug, Metadata: []any{}}}})
			if err != nil {
				r.fail(d, "scope binding", err)
				return m, false
			}
			scope, err = r.client.sdk.IAMScopes.Get(ctx, name)
			if err != nil {
				r.fail(d, "scope binding confirmation", err)
				return m, false
			}
		}
		if !dedicatedScope(scope, name, slug, r.client.organisation, true, d) {
			return m, false
		}
	} else if scope.Name != name || scope.TSGID != r.client.organisation {
		d.AddError("Unexpected external IAM scope", "The scope name and tenant must match the configured provider.")
		return m, false
	}
	bound := false
	for _, binding := range scope.Resources {
		if binding.ResourceType == "workspace" && binding.ResourceID == slug {
			bound = true
		}
	}
	return workspaceChange(m, "scope_binding_ready", bound), true
}
func (r *workspaceResource) provision(ctx context.Context, m types.Object, state *tfsdk.State, d *diag.Diagnostics, applied bool) (types.Object, bool) {
	stage := workspaceValue(m, "provisioning_stage")
	if stage == "scope_create_uncertain" {
		var ok bool
		m, ok = r.recoverScope(ctx, m, d)
		if !ok {
			if !d.HasError() {
				d.AddError("Scope creation outcome is unresolved", "The uncertain scope POST is not replayed. Its token and pending identity remain in state; confirm the remote outcome before continuing.")
			}
			return m, false
		}
		if !workspaceCheckpoint(ctx, state, m, d) {
			return m, false
		}
		stage = workspaceValue(m, "provisioning_stage")
	}
	if stage == "workspace_create_uncertain" {
		var ok bool
		m, ok = r.reconcileWorkspace(ctx, m, d)
		if !workspaceCheckpoint(ctx, state, m, d) || !ok {
			return m, false
		}
		stage = workspaceValue(m, "provisioning_stage")
	}
	if stage == "scope_ready" {
		q, err := r.validateCreate(m)
		if err != nil {
			r.fail(d, "validation", err)
			return m, false
		}
		m = workspaceChange(m, "provisioning_stage", "workspace_create_uncertain")
		if !workspaceCheckpoint(ctx, state, m, d) {
			return m, false
		}
		receipt, err := r.client.sdk.Workspaces.Create(ctx, q)
		if err != nil {
			if workspaceRejected(err) {
				m = workspaceChange(m, "provisioning_stage", "scope_ready")
				workspaceCheckpoint(ctx, state, m, d)
			}
			r.fail(d, "creation", err)
			return m, false
		}
		m = workspaceChange(m, "id", receipt.ID)
		m = workspaceChange(m, "slug", receipt.Slug)
		m = workspaceChange(m, "provisioning_stage", "workspace_created")
		if !workspaceCheckpoint(ctx, state, m, d) {
			return m, false
		}
	}
	if workspaceValue(m, "provisioning_stage") == "archived" {
		d.AddError("Workspace is archived", "Plan replacement to clean the recorded owned scope and provision a new workspace.")
		return m, false
	}
	var ok bool
	m, ok = r.bindScope(ctx, m, d)
	if !ok {
		return m, false
	}
	if !workspaceCheckpoint(ctx, state, m, d) {
		return m, false
	}
	remote, err := r.detail(ctx, workspaceValue(m, "id"))
	if err != nil {
		r.fail(d, "read after provisioning", err)
		return m, false
	}
	if err = workspaceIdentity(m, remote, applied); err != nil {
		r.fail(d, "identity confirmation", err)
		return m, false
	}
	m, err = workspaceReadSettings(ctx, m, remote, applied)
	if err != nil {
		r.fail(d, "settings confirmation", err)
		return m, false
	}
	m = workspaceReady(workspaceState(m, remote, applied))
	return m, workspaceCheckpoint(ctx, state, m, d)
}
func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var m types.Object
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.validateCreate(m); err != nil {
		r.fail(&resp.Diagnostics, "validation", err)
		return
	}
	m = workspaceInitial(m)
	scope := workspaceValue(m, "scope_name")
	managed := workspaceValue(m, "scope_management") == "managed"
	existing, err := r.client.sdk.IAMScopes.Get(ctx, scope)
	if managed {
		if err == nil {
			resp.Diagnostics.AddError("IAM scope already exists", "Use a unique scope_name or explicitly import the dedicated scope/workspace relationship. Existing scopes are never automatically adopted.")
			return
		}
		if !aisec.IsNotFound(err) {
			r.fail(&resp.Diagnostics, "scope preflight", err)
			return
		}
		token := make([]byte, 16)
		if _, err = rand.Read(token); err != nil {
			r.fail(&resp.Diagnostics, "scope intent", err)
			return
		}
		m = workspaceChange(m, "scope_ownership_token", hex.EncodeToString(token))
		m = workspaceChange(m, "id", "pending:"+scope)
		m = workspaceChange(m, "provisioning_stage", "scope_create_uncertain")
		if !workspaceCheckpoint(ctx, &resp.State, m, &resp.Diagnostics) {
			return
		}
		receipt, err := r.client.sdk.IAMScopes.Create(ctx, gw.IAMScopeCreateInput{Name: scope, Description: workspaceScopeDescription(m)})
		if err != nil {
			if workspaceRejected(err) {
				resp.State.RemoveResource(ctx)
			}
			r.fail(&resp.Diagnostics, "scope creation", err)
			return
		}
		if !dedicatedScope(receipt, scope, "", r.client.organisation, false, &resp.Diagnostics) {
			return
		}
		m = workspaceChange(m, "scope_owned", true)
	} else {
		if err != nil {
			r.fail(&resp.Diagnostics, "external scope read", err)
			return
		}
		if existing.Name != scope || existing.TSGID != r.client.organisation {
			resp.Diagnostics.AddError("Unexpected external IAM scope", "The scope name and tenant must match the configured provider.")
			return
		}
		m = workspaceChange(m, "id", "pending:"+scope)
	}
	m = workspaceChange(m, "provisioning_stage", "scope_ready")
	if !workspaceCheckpoint(ctx, &resp.State, m, &resp.Diagnostics) {
		return
	}
	r.provision(ctx, m, &resp.State, &resp.Diagnostics, true)
}
func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var m types.Object
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if workspaceValue(m, "provisioning_stage") == "scope_create_uncertain" {
		recovered, ok := r.recoverScope(ctx, m, &resp.Diagnostics)
		if ok {
			workspaceCheckpoint(ctx, &resp.State, recovered, &resp.Diagnostics)
		}
		return
	}
	if strings.HasPrefix(workspaceValue(m, "id"), "pending:") {
		if workspaceValue(m, "provisioning_stage") == "workspace_create_uncertain" {
			recovered, ok := r.reconcileWorkspace(ctx, m, &resp.Diagnostics)
			if ok {
				workspaceCheckpoint(ctx, &resp.State, recovered, &resp.Diagnostics)
			}
		}
		return
	}
	remote, err := r.detail(ctx, workspaceValue(m, "id"))
	if aisec.IsNotFound(err) {
		m = workspaceChange(m, "status", "archived")
		m = workspaceChange(m, "provisioning_stage", "archived")
		workspaceCheckpoint(ctx, &resp.State, m, &resp.Diagnostics)
		return
	}
	if err != nil {
		r.fail(&resp.Diagnostics, "refresh", err)
		return
	}
	if err = workspaceIdentity(m, remote, false); err != nil {
		r.fail(&resp.Diagnostics, "identity confirmation", err)
		return
	}
	m, err = workspaceReadSettings(ctx, m, remote, false)
	if err != nil {
		r.fail(&resp.Diagnostics, "settings refresh", err)
		return
	}
	m = workspaceState(m, remote, false)
	scope, err := r.client.sdk.IAMScopes.Get(ctx, workspaceValue(m, "scope_name"))
	if aisec.IsNotFound(err) {
		m = workspaceChange(m, "provisioning_stage", "scope_missing")
		m = workspaceChange(m, "scope_binding_ready", false)
		if !workspaceFlag(m, "scope_owned") {
			resp.Diagnostics.AddWarning("External IAM scope is missing", "The workspace remains managed at its existing UUID. Restore the externally managed scope before applying updates; Terraform will not replace the workspace or write IAM. Explicit destruction still archives the workspace.")
		}
	} else if err != nil {
		r.fail(&resp.Diagnostics, "scope refresh", err)
		return
	} else {
		if scope.Name != workspaceValue(m, "scope_name") || scope.TSGID != r.client.organisation {
			resp.Diagnostics.AddError("Unexpected IAM scope identity", "Returned scope name and tenant differ from recorded identity.")
			return
		}
		if workspaceFlag(m, "scope_owned") && !dedicatedScope(scope, workspaceValue(m, "scope_name"), workspaceValue(m, "slug"), r.client.organisation, false, &resp.Diagnostics) {
			return
		}
		bound := false
		for _, binding := range scope.Resources {
			if binding.ResourceType == "workspace" && binding.ResourceID == workspaceValue(m, "slug") {
				bound = true
			}
		}
		m = workspaceChange(m, "scope_binding_ready", bound)
		if workspaceValue(m, "provisioning_stage") == "scope_missing" {
			m = workspaceChange(m, "provisioning_stage", "ready")
		}
		if workspaceFlag(m, "scope_owned") && !bound {
			m = workspaceChange(m, "provisioning_stage", "workspace_created")
		}
	}
	workspaceCheckpoint(ctx, &resp.State, m, &resp.Diagnostics)
}
func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var m, old types.Object
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a := m.Attributes()
	for _, key := range workspaceComputedFields() {
		a[key] = old.Attributes()[key]
	}
	m = types.ObjectValueMust(m.AttributeTypes(ctx), a)
	baseline := old
	if workspaceValue(old, "provisioning_stage") != "ready" {
		resume := old
		if stage := workspaceValue(old, "provisioning_stage"); stage == "scope_ready" || stage == "scope_create_uncertain" {
			resume = m
		}
		recovered, ok := r.provision(ctx, resume, &resp.State, &resp.Diagnostics, false)
		if !ok {
			return
		}
		baseline = recovered
		a = m.Attributes()
		for _, key := range workspaceComputedFields() {
			a[key] = recovered.Attributes()[key]
		}
		m = types.ObjectValueMust(m.AttributeTypes(ctx), a)
	}
	b, clear, err := workspaceSettingsChanges(m, baseline)
	if err != nil {
		r.fail(&resp.Diagnostics, "validation", err)
		return
	}
	// Validate the entire desired write before any clearing mutation.
	var update gw.WorkspaceUpdateRequest
	if len(b) > 0 {
		update, err = workspaceDecode[gw.WorkspaceUpdateRequest](b)
		if err != nil {
			r.fail(&resp.Diagnostics, "validation", &inputShapeError{field: "settings", detail: "do not match the workspace contract"})
			return
		}
		if err = parity.Validate("GatewayWorkspaceUpdateRequestSchema", update); err != nil {
			r.fail(&resp.Diagnostics, "validation", &inputShapeError{field: "settings", detail: "do not satisfy the workspace contract"})
			return
		}
	}
	if workspaceAnyClear(clear) {
		if err = r.client.sdk.Workspaces.ClearSettings(ctx, workspaceValue(m, "id"), clear); err != nil {
			r.fail(&resp.Diagnostics, "settings clearing", err)
			return
		}
	}
	if len(b) > 0 {
		if err = r.client.sdk.Workspaces.Update(ctx, workspaceValue(m, "id"), update); err != nil {
			r.fail(&resp.Diagnostics, "update", err)
			return
		}
	}
	remote, err := r.detail(ctx, workspaceValue(m, "id"))
	if err != nil {
		r.fail(&resp.Diagnostics, "read after update", err)
		return
	}
	if err = workspaceConfirmRemovals(m, remote, clear); err != nil {
		r.fail(&resp.Diagnostics, "clearing confirmation", err)
		return
	}
	if err = workspaceIdentity(m, remote, true); err != nil {
		r.fail(&resp.Diagnostics, "identity confirmation", err)
		return
	}
	m, err = workspaceReadSettings(ctx, m, remote, true)
	if err != nil {
		r.fail(&resp.Diagnostics, "settings confirmation", err)
		return
	}
	workspaceCheckpoint(ctx, &resp.State, workspaceReady(workspaceState(m, remote, true)), &resp.Diagnostics)
}
func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var m types.Object
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if workspaceValue(m, "provisioning_stage") == "scope_create_uncertain" {
		recovered, ok := r.recoverScope(ctx, m, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		if !ok {
			resp.State.RemoveResource(ctx)
			return
		}
		m = recovered
		if !workspaceCheckpoint(ctx, &resp.State, m, &resp.Diagnostics) {
			return
		}
	}
	if workspaceValue(m, "provisioning_stage") == "workspace_create_uncertain" {
		recovered, ok := r.reconcileWorkspace(ctx, m, &resp.Diagnostics)
		m = recovered
		if !ok || !workspaceCheckpoint(ctx, &resp.State, m, &resp.Diagnostics) {
			return
		}
	}
	if workspaceFlag(m, "scope_owned") && !r.checkOwnedScope(ctx, m, &resp.Diagnostics) {
		return
	}
	id := workspaceValue(m, "id")
	if !strings.HasPrefix(id, "pending:") {
		remote, err := r.detail(ctx, id)
		if err != nil && !aisec.IsNotFound(err) {
			r.fail(&resp.Diagnostics, "archive preflight", err)
			return
		}
		if err == nil {
			if err = workspaceIdentity(m, remote, false); err != nil {
				r.fail(&resp.Diagnostics, "archive identity confirmation", err)
				return
			}
			if err = r.client.sdk.Workspaces.Delete(ctx, id); err != nil {
				r.fail(&resp.Diagnostics, "archive", err)
				return
			}
		}
		_, err = r.detail(ctx, id)
		if !aisec.IsNotFound(err) {
			if err == nil {
				err = fmt.Errorf("workspace remains active")
			}
			r.fail(&resp.Diagnostics, "archive confirmation", err)
			return
		}
	}
	m = workspaceChange(m, "status", "archived")
	m = workspaceChange(m, "provisioning_stage", "archived")
	if !workspaceCheckpoint(ctx, &resp.State, m, &resp.Diagnostics) {
		return
	}
	if workspaceFlag(m, "scope_owned") {
		scope := workspaceValue(m, "scope_name")
		record, err := r.client.sdk.IAMScopes.Get(ctx, scope)
		if !aisec.IsNotFound(err) {
			if err != nil {
				r.fail(&resp.Diagnostics, "scope cleanup preflight", err)
				return
			}
			if !dedicatedScope(record, scope, workspaceValue(m, "slug"), r.client.organisation, false, &resp.Diagnostics) {
				return
			}
			if err = r.client.sdk.IAMScopes.Delete(ctx, scope); err != nil {
				r.fail(&resp.Diagnostics, "scope cleanup", err)
				return
			}
		}
		_, err = r.client.sdk.IAMScopes.Get(ctx, scope)
		if !aisec.IsNotFound(err) {
			if err == nil {
				err = fmt.Errorf("scope remains")
			}
			r.fail(&resp.Diagnostics, "scope cleanup confirmation", err)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *workspaceResource) checkOwnedScope(ctx context.Context, m types.Object, d *diag.Diagnostics) bool {
	scope, err := r.client.sdk.IAMScopes.Get(ctx, workspaceValue(m, "scope_name"))
	if aisec.IsNotFound(err) {
		return true
	}
	if err != nil {
		r.fail(d, "scope ownership check", err)
		return false
	}
	return dedicatedScope(scope, workspaceValue(m, "scope_name"), workspaceValue(m, "slug"), r.client.organisation, false, d)
}

func workspaceIdentity(m types.Object, remote document, applied bool) error {
	for _, key := range []string{"id", "slug"} {
		value, ok := remote[key].(string)
		if !ok || value == "" || value != workspaceValue(m, key) {
			return &inputShapeError{field: key, detail: "does not match the recorded workspace identity"}
		}
	}
	if value, ok := remote["scope_name"].(string); ok && value != "" && value != workspaceValue(m, "scope_name") {
		return &inputShapeError{field: "scope_name", detail: "does not match the recorded workspace relationship"}
	}
	if applied && workspaceLabel(remote) != workspaceValue(m, "name") {
		return &inputShapeError{field: "name", detail: "was not confirmed after the write"}
	}
	return nil
}

// A typed, explicit client rejection is distinct from timeout, throttling,
// transport failure or undecodable success, whose POST outcome remains uncertain.
func workspaceRejected(err error) bool {
	var sdkErr *aisec.AISecSDKError
	return errors.As(err, &sdkErr) && sdkErr.StatusCode >= 400 && sdkErr.StatusCode < 500 && sdkErr.StatusCode != 408 && sdkErr.StatusCode != 425 && sdkErr.StatusCode != 429
}

// Admin detail decorates the label with the icon; list rows use the plain label.
// Remove precisely one matching decoration, preserving any literal icon prefix
// already present in the caller's actual label.
func workspaceLabel(remote document) string {
	name, _ := remote["name"].(string)
	icon, _ := remote["icon"].(string)
	if icon != "" {
		return strings.TrimPrefix(name, icon+" ")
	}
	return name
}

func workspaceComputedFields() []string {
	return []string{"id", "slug", "status", "created_at", "last_updated_at", "scope_owned", "scope_ownership_token", "scope_binding_ready", "provisioning_stage"}
}
