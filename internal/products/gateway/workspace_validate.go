package gateway

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithValidateConfig = &workspaceResource{}

func workspaceUnknown(v attr.Value) bool {
	if v.IsUnknown() {
		return true
	}
	if v.IsNull() {
		return false
	}
	switch x := v.(type) {
	case types.Dynamic:
		return workspaceUnknown(x.UnderlyingValue())
	case types.Object:
		for _, value := range x.Attributes() {
			if workspaceUnknown(value) {
				return true
			}
		}
	case types.List:
		for _, value := range x.Elements() {
			if workspaceUnknown(value) {
				return true
			}
		}
	case types.Map:
		for _, value := range x.Elements() {
			if workspaceUnknown(value) {
				return true
			}
		}
	case types.Set:
		for _, value := range x.Elements() {
			if workspaceUnknown(value) {
				return true
			}
		}
	case types.Tuple:
		for _, value := range x.Elements() {
			if workspaceUnknown(value) {
				return true
			}
		}
	}
	return false
}
func (r *workspaceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, key := range []string{"name", "description"} {
		v := m.Attributes()[key]
		if !v.IsNull() && !v.IsUnknown() && strings.TrimSpace(v.(types.String).ValueString()) == "" {
			resp.Diagnostics.AddError("Blank workspace label", "Workspace name and configured description must contain non-whitespace characters. The API does not clear blank descriptions.")
		}
	}
	for _, key := range []string{"defaults", "usage_limits", "rate_limits"} {
		v := m.Attributes()[key]
		if v.IsNull() || workspaceUnknown(v) {
			continue
		}
		if _, err := workspaceInput(v); err != nil {
			resp.Diagnostics.AddError("Invalid workspace settings", gatewayError(err))
		}
	}
	list := m.Attributes()["usage_limits"].(types.List)
	if list.IsNull() || workspaceUnknown(list) {
		return
	}
	seen := map[string]bool{}
	for _, item := range list.Elements() {
		fields := item.(types.Object).Attributes()
		kind := fields["type"].(types.String).ValueString()
		if seen[kind] {
			resp.Diagnostics.AddError("Duplicate workspace usage type", "Each measurement type may appear only once in the owned usage-policy collection.")
		}
		seen[kind] = true
		if !fields["periodic_reset"].IsNull() && !fields["periodic_reset_days"].IsNull() {
			resp.Diagnostics.AddError("Conflicting workspace reset settings", "Choose periodic_reset or periodic_reset_days, rather than both.")
		}
		if v := fields["next_usage_reset_at"]; !v.IsNull() {
			if _, err := time.Parse(time.RFC3339Nano, v.(types.String).ValueString()); err != nil {
				resp.Diagnostics.AddError("Invalid workspace reset timestamp", "next_usage_reset_at must be an RFC 3339 timestamp.")
			}
		}
	}
}
