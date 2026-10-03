package gateway

import (
	"context"
	"fmt"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func workspaceSettingsAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"defaults": schema.SingleNestedAttribute{Optional: true, Description: "Owns the complete default config selection and metadata object. Removing a previously configured object clears those fields. Unreadable allow_config_override is not supported.", Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{Optional: true, Description: "Existing default config UUID. Use references rather than embedded credentials.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"metadata":  schema.DynamicAttribute{Optional: true, Description: "Native HCL metadata object. Embedded credential fields are rejected."},
		}},
		"rate_limits": schema.ListNestedAttribute{Optional: true, Description: "Owns the complete workspace rate-policy collection. Empty or removed collections clear the previously managed policies.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"type":  schema.StringAttribute{Required: true, Description: "Measurement type, for example requests.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"unit":  schema.StringAttribute{Required: true, Description: "Rate unit, for example rpm.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"value": schema.Int64Attribute{Required: true, Description: "Nonnegative integer rate.", Validators: []validator.Int64{int64validator.AtLeast(0)}},
		}}},
		"usage_limits": schema.ListNestedAttribute{Optional: true, Validators: []validator.List{listvalidator.SizeAtMost(1)}, Description: "Owns the complete workspace usage-policy collection (the service permits one policy). Empty or removed collections clear its associations; hidden backing-record deletion is not asserted.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"type":                schema.StringAttribute{Required: true, Description: "Usage measurement type, for example tokens.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"credit_limit":        schema.Float64Attribute{Required: true, Description: "Nonnegative usage allowance.", Validators: []validator.Float64{float64validator.AtLeast(0)}},
			"alert_threshold":     schema.Float64Attribute{Optional: true, Description: "Optional nonnegative alert threshold; zero is explicit.", Validators: []validator.Float64{float64validator.AtLeast(0)}},
			"periodic_reset":      schema.StringAttribute{Optional: true, Description: "Optional reset cadence, such as monthly or weekly."},
			"periodic_reset_days": schema.Int64Attribute{Optional: true, Description: "Optional positive custom reset interval in days.", Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"next_usage_reset_at": schema.StringAttribute{Optional: true, Description: "Optional next reset timestamp in RFC 3339 format."},
		}}},
	}
}

// Nil object fields are omitted at the typed SDK seam. The separately typed clear
// helper handles fields for which omission and explicit null have different meaning.
func workspaceInput(v attr.Value) (any, error) {
	x, err := nativeJSON(v)
	if err != nil {
		return nil, err
	}
	if _, ok := v.(types.Object); ok {
		body := x.(map[string]any)
		for k, value := range body {
			if value == nil {
				delete(body, k)
			}
		}
		if metadata, present := body["metadata"]; present {
			if _, ok := metadata.(map[string]any); !ok {
				return nil, &inputShapeError{field: "defaults.metadata", detail: "must be a native HCL object"}
			}
			if routingCredentials(metadata) {
				return nil, &inputShapeError{field: "defaults.metadata", detail: "must use credential references, never plaintext credential fields"}
			}
		}
		return body, nil
	}
	if list, ok := v.(types.List); ok {
		result := make([]any, len(list.Elements()))
		for i, item := range list.Elements() {
			value, err := workspaceInput(item)
			if err != nil {
				return nil, err
			}
			result[i] = value
		}
		return result, nil
	}
	return x, nil
}
func workspaceReadSettings(ctx context.Context, m types.Object, remote document, applied bool) (types.Object, error) {
	values := m.Attributes()
	for _, key := range []string{"description", "icon", "defaults", "rate_limits", "usage_limits"} {
		prior := values[key]
		if prior.IsNull() {
			continue
		} // settings never configured remain unmanaged
		value := remote[key]
		var mapped attr.Value
		var err error
		switch p := prior.(type) {
		case types.Object:
			doc, ok := value.(map[string]any)
			if !ok {
				doc = map[string]any{}
			}
			if metadata := doc["metadata"]; routingCredentials(metadata) {
				return m, &inputShapeError{field: "defaults.metadata", detail: "contains an embedded credential remotely; remove it or stop managing defaults"}
			}
			// An omitted metadata object is equivalent to an empty remote metadata object.
			if metadata, ok := doc["metadata"].(map[string]any); ok && len(metadata) == 0 && p.Attributes()["metadata"].IsNull() {
				doc["metadata"] = nil
			}
			mapped, err = readTypedObject(ctx, doc, p)
		case types.List:
			rows, ok := value.([]any)
			if value == nil {
				rows = []any{}
			} else if !ok {
				return m, fmt.Errorf("invalid workspace policy collection")
			}
			items := make([]attr.Value, len(rows))
			templateType := p.ElementType(ctx).(types.ObjectType)
			for i, row := range rows {
				template := types.ObjectNull(templateType.AttrTypes)
				if i < len(p.Elements()) {
					template = p.Elements()[i].(types.Object)
				}
				items[i], err = readTypedObject(ctx, row, template)
				if err != nil {
					return m, err
				}
			}
			list, d := types.ListValue(p.ElementType(ctx), items)
			if d.HasError() {
				return m, fmt.Errorf("cannot read workspace policies")
			}
			mapped = list
		default:
			if value == nil && key == "icon" && p.(types.String).ValueString() == "" {
				value = ""
			}
			mapped, err = readNative(ctx, value, prior)
		}
		if err != nil {
			return m, err
		}
		if applied && !mapped.Equal(prior) {
			return m, &inputShapeError{field: key, detail: "was not confirmed after the write; refresh will reconcile the remote value before retrying"}
		}
		values[key] = mapped
	}
	return types.ObjectValueMust(m.AttributeTypes(ctx), values), nil
}

// Workspace usage writes update the one existing policy by measurement type.
// Credit changes preserve its UUID. Type changes and removal of nonnullable
// optional fields require clearing the old association before configuring anew.
func workspaceSettingsChanges(m, old types.Object) (document, gw.WorkspaceClearSettingsRequest, error) {
	body, err := workspacePayload(m, false)
	clear := gw.WorkspaceClearSettingsRequest{}
	if err != nil {
		return nil, clear, err
	}
	for _, key := range []string{"name", "description", "icon", "defaults", "usage_limits", "rate_limits"} {
		wanted, previous := m.Attributes()[key], old.Attributes()[key]
		if wanted.Equal(previous) {
			delete(body, key)
			continue
		}
		switch key {
		case "icon":
			if wanted.IsNull() && !previous.IsNull() {
				clear.Icon = true
			}
		case "defaults":
			if wanted.IsNull() {
				if !previous.IsNull() {
					clear.Defaults = true
				}
				continue
			}
			fields := wanted.(types.Object).Attributes()
			if fields["config_id"].IsNull() {
				clear.Defaults = true
			}
			defaults := body[key].(map[string]any)
			if fields["metadata"].IsNull() {
				defaults["metadata"] = map[string]any{}
			}
		case "usage_limits":
			if wanted.IsNull() || len(wanted.(types.List).Elements()) == 0 {
				clear.UsageLimits = !previous.IsNull() || !wanted.IsNull()
				delete(body, key)
				continue
			}
			if previous.IsNull() || len(previous.(types.List).Elements()) == 0 {
				clear.UsageLimits = true
				continue
			}
			before := previous.(types.List).Elements()[0].(types.Object).Attributes()
			after := wanted.(types.List).Elements()[0].(types.Object).Attributes()
			if !before["type"].Equal(after["type"]) {
				clear.UsageLimits = true
			}
			for _, field := range []string{"alert_threshold", "periodic_reset", "periodic_reset_days", "next_usage_reset_at"} {
				if after[field].IsNull() && !before[field].IsNull() {
					clear.UsageLimits = true
				}
			}
		case "rate_limits":
			if wanted.IsNull() && !previous.IsNull() {
				clear.RateLimits = true
			}
		}
	}
	return body, clear, nil
}
func workspaceAnyClear(q gw.WorkspaceClearSettingsRequest) bool {
	return q.Defaults || q.UsageLimits || q.RateLimits || q.Icon
}

func workspaceConfirmRemovals(m types.Object, remote document, clear gw.WorkspaceClearSettingsRequest) error {
	bad := func(field string) error {
		return &inputShapeError{field: field, detail: "was acknowledged but its removal was not confirmed; prior ownership remains in state"}
	}
	if clear.Icon && m.Attributes()["icon"].IsNull() {
		if v := remote["icon"]; v != nil && v != "" {
			return bad("icon")
		}
	}
	if clear.Defaults && m.Attributes()["defaults"].IsNull() {
		d, _ := remote["defaults"].(map[string]any)
		metadata, _ := d["metadata"].(map[string]any)
		if d["config_id"] != nil || len(metadata) > 0 {
			return bad("defaults")
		}
	}
	for key, selected := range map[string]bool{"usage_limits": clear.UsageLimits, "rate_limits": clear.RateLimits} {
		if !selected {
			continue
		}
		wanted := m.Attributes()[key]
		if !wanted.IsNull() && len(wanted.(types.List).Elements()) > 0 {
			continue
		}
		value := remote[key]
		if value == nil {
			continue
		}
		rows, ok := value.([]any)
		if !ok || len(rows) > 0 {
			return bad(key)
		}
	}
	return nil
}
