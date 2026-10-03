package gateway

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithModifyPlan = &workspaceResource{}

func (r *workspaceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	var old types.Object
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status := workspaceValue(old, "status")
	scopeMissing := workspaceValue(old, "provisioning_stage") == "scope_missing"
	if scopeMissing && !workspaceFlag(old, "scope_owned") {
		resp.Diagnostics.AddError("External IAM scope is missing", "Restore the externally managed scope before applying updates. The existing workspace is retained; Terraform will not replace it or write IAM. An explicit destroy can still archive the workspace.")
		return
	}
	if status == "archived" || (scopeMissing && workspaceFlag(old, "scope_owned")) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("id"))
		for _, key := range []string{"id", "slug", "status", "created_at", "last_updated_at", "scope_ownership_token", "provisioning_stage"} {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(key), types.StringUnknown())...)
		}
		for _, key := range []string{"scope_owned", "scope_binding_ready"} {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(key), types.BoolUnknown())...)
		}
		return
	}
	if workspaceValue(old, "provisioning_stage") != "ready" {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("scope_owned"), types.BoolUnknown())...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("provisioning_stage"), types.StringUnknown())...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("scope_binding_ready"), types.BoolUnknown())...)
		if strings.HasPrefix(workspaceValue(old, "id"), "pending:") {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("slug"), types.StringUnknown())...)
		}
	}
}
