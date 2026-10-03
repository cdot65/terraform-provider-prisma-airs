package gateway

import (
	"context"
	"strings"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithImportState = &workspaceResource{}

func dedicatedScope(scope *gw.IAMScope, name, slug, tenant string, requireBinding bool, d *diag.Diagnostics) bool {
	if scope.Name != name || scope.TSGID != tenant {
		d.AddError("Unexpected IAM scope identity", "Scope name and tenant must match recorded ownership.")
		return false
	}
	found := false
	for _, binding := range scope.Resources {
		if binding.ResourceType != "workspace" || slug == "" || binding.ResourceID != slug {
			d.AddError("IAM scope is shared", "The dedicated scope contains an unrelated resource binding. Resolve ownership explicitly; no IAM changes are allowed.")
			return false
		}
		found = true
	}
	if requireBinding && !found {
		d.AddError("Workspace binding is missing", "Explicit adoption requires a dedicated scope already bound to this workspace slug.")
		return false
	}
	return true
}
func (r *workspaceResource) emptyModel(ctx context.Context) types.Object {
	var s resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &s)
	ts := map[string]attr.Type{}
	values := map[string]attr.Value{}
	for k, a := range s.Schema.Attributes {
		ts[k] = a.GetType()
		values[k], _ = nullNative(ctx, ts[k])
	}
	return workspaceInitial(types.ObjectValueMust(ts, values))
}
func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	managed := len(parts) == 3 && parts[0] == "managed"
	if managed {
		parts = parts[1:]
	}
	if len(parts) < 1 || len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		resp.Diagnostics.AddError("Invalid workspace import identity", "Use <workspace-uuid>/<scope-name> for external ownership, or managed/<workspace-uuid>/<scope-name> for explicit dedicated-scope adoption.")
		return
	}
	remote, e := r.detail(ctx, parts[0])
	if e != nil {
		r.fail(&resp.Diagnostics, "import lookup", e)
		return
	}
	scopeName, _ := remote["scope_name"].(string)
	if len(parts) == 2 {
		if scopeName != "" && scopeName != parts[1] {
			resp.Diagnostics.AddError("Workspace scope mismatch", "The import scope does not match the workspace's recorded scope_name.")
			return
		}
		scopeName = parts[1]
	}
	if scopeName == "" {
		resp.Diagnostics.AddError("Scope name is required", "The workspace response omitted scope_name. Supply <workspace-uuid>/<scope-name>; ownership cannot be inferred.")
		return
	}
	scope, e := r.client.sdk.IAMScopes.Get(ctx, scopeName)
	if e != nil {
		r.fail(&resp.Diagnostics, "import scope lookup", e)
		return
	}
	slug, _ := remote["slug"].(string)
	if scope.Name != scopeName || scope.TSGID != r.client.organisation {
		resp.Diagnostics.AddError("Unexpected imported scope", "The scope's name and tenant must match this provider.")
		return
	}
	if managed && !dedicatedScope(scope, scopeName, slug, r.client.organisation, true, &resp.Diagnostics) {
		return
	}
	m := r.emptyModel(ctx)
	m = workspaceChange(m, "scope_name", scopeName)
	mode := "external"
	if managed {
		mode = "managed"
	}
	m = workspaceChange(m, "scope_management", mode)
	m = workspaceChange(m, "scope_owned", managed)
	bound := false
	for _, binding := range scope.Resources {
		if binding.ResourceType == "workspace" && binding.ResourceID == slug {
			bound = true
		}
	}
	m = workspaceChange(m, "scope_binding_ready", bound)
	resp.Diagnostics.Append(resp.State.Set(ctx, workspaceReady(workspaceState(m, remote, false)))...)
}
