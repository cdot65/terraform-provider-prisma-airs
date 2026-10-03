package gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func checksAttribute(description string) schema.Attribute {
	return schema.ListNestedAttribute{Required: true, Description: description, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Required: true, Description: "Check identifier.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"name":       schema.StringAttribute{Optional: true, Computed: true, Description: "Optional check name."},
		"is_enabled": schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the check is enabled; false is explicit."},
	}}}
}

func actionsAttribute(description string) schema.Attribute {
	feedback := func() schema.Attribute {
		return schema.SingleNestedAttribute{Required: true, Attributes: map[string]schema.Attribute{"value": schema.Float64Attribute{Required: true, Description: "Feedback value."}, "weight": schema.Float64Attribute{Required: true, Description: "Feedback weight."}, "metadata": schema.StringAttribute{Required: true, Description: "Feedback metadata; an empty string is explicit."}}}
	}
	return schema.SingleNestedAttribute{Required: true, Description: description, Attributes: map[string]schema.Attribute{
		"deny": schema.BoolAttribute{Required: true, Description: "Deny on failure."}, "async": schema.BoolAttribute{Required: true, Description: "Evaluate asynchronously."},
		"on_success": schema.SingleNestedAttribute{Required: true, Attributes: map[string]schema.Attribute{"feedback": feedback()}},
		"on_fail":    schema.SingleNestedAttribute{Required: true, Attributes: map[string]schema.Attribute{"feedback": feedback()}},
	}}
}

func readTypedObject(ctx context.Context, value any, prior types.Object) (attr.Value, error) {
	remote, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid typed object response")
	}
	ts := prior.AttributeTypes(ctx)
	values := map[string]attr.Value{}
	for k, t := range ts {
		template := prior.Attributes()[k]
		if template == nil {
			var e error
			template, e = nullNative(ctx, t)
			if e != nil {
				return nil, e
			}
		}
		v, present := remote[k]
		if !present {
			v = nil
		}
		var mapped attr.Value
		var e error
		if o, ok := template.(types.Object); ok && v != nil {
			mapped, e = readTypedObject(ctx, v, o)
		} else {
			mapped, e = readNative(ctx, v, template)
		}
		if e != nil {
			return nil, e
		}
		values[k] = mapped
	}
	v, d := types.ObjectValue(ts, values)
	if d.HasError() {
		return nil, fmt.Errorf("cannot read typed object")
	}
	return v, nil
}

func writeChecks(v types.List) ([]any, error) {
	checks := make([]any, len(v.Elements()))
	for i, value := range v.Elements() {
		o := value.(types.Object)
		body := map[string]any{}
		for k, x := range o.Attributes() {
			if x.IsNull() {
				continue
			}
			if x.IsUnknown() && k != "id" {
				continue
			}
			encoded, e := nativeJSON(x)
			if e != nil {
				return nil, e
			}
			body[k] = encoded
		}
		checks[i] = body
	}
	return checks, nil
}

// Check identities and flags are readable. Parameters are separately retained
// as sensitive desired input and are never reconstructed from masked GET values.
func readChecks(ctx context.Context, value any, prior types.List) (attr.Value, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid checks response")
	}
	ts := prior.ElementType(ctx).(types.ObjectType).AttrTypes
	old := map[string]types.Object{}
	for _, v := range prior.Elements() {
		o := v.(types.Object)
		old[o.Attributes()["id"].(types.String).ValueString()] = o
	}
	values := make([]attr.Value, len(items))
	for i, item := range items {
		remote, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid check response")
		}
		id, ok := remote["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("check response omitted its identity")
		}
		attrs := map[string]attr.Value{}
		for k, t := range ts {
			v, e := nullNative(ctx, t)
			if e != nil {
				return nil, e
			}
			attrs[k] = v
		}
		attrs["id"] = types.StringValue(id)
		for _, k := range []string{"name", "is_enabled"} {
			if v, present := remote[k]; present {
				mapped, e := readNative(ctx, v, attrs[k])
				if e != nil {
					return nil, e
				}
				attrs[k] = mapped
			}
		}
		if o, present := old[id]; present {
			for _, k := range []string{"name", "is_enabled"} {
				if _, present := remote[k]; !present && !o.Attributes()[k].IsUnknown() {
					attrs[k] = o.Attributes()[k]
				}
			}
		}
		v, d := types.ObjectValue(ts, attrs)
		if d.HasError() {
			return nil, fmt.Errorf("cannot read check state")
		}
		values[i] = v
	}
	v, d := types.ListValue(prior.ElementType(ctx), values)
	if d.HasError() {
		return nil, fmt.Errorf("cannot read check list")
	}
	return v, nil
}
