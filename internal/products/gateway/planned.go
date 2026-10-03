package gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Keep the planned shape and every known leaf on apply. Only Read exposes
// additional remote fields; unknown children are filled without changing types.
func honorPlanned(plan, remote attr.Value) (attr.Value, error) {
	ctx := context.Background()
	if plan.IsUnknown() {
		if remote == nil {
			return nullNative(ctx, plan.Type(ctx))
		}
		value, err := remote.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}
		return plan.Type(ctx).ValueFromTerraform(ctx, value)
	}
	if _, err := nativeJSON(plan); err == nil {
		return plan, nil
	}
	if p, ok := plan.(types.Dynamic); ok {
		var value attr.Value
		if r, ok := remote.(types.Dynamic); ok {
			value = r.UnderlyingValue()
		} else {
			value = remote
		}
		mapped, err := honorPlanned(p.UnderlyingValue(), value)
		if err != nil {
			return nil, err
		}
		return types.DynamicValue(mapped), nil
	}
	switch p := plan.(type) {
	case types.Object:
		old := map[string]attr.Value{}
		switch r := remote.(type) {
		case types.Object:
			old = r.Attributes()
		case types.Map:
			old = r.Elements()
		}
		values := map[string]attr.Value{}
		for k, v := range p.Attributes() {
			mapped, err := honorPlanned(v, old[k])
			if err != nil {
				return nil, err
			}
			values[k] = mapped
		}
		value, d := types.ObjectValue(p.AttributeTypes(ctx), values)
		if d.HasError() {
			return nil, fmt.Errorf("incompatible planned object")
		}
		return value, nil
	case types.Map:
		old := map[string]attr.Value{}
		switch r := remote.(type) {
		case types.Object:
			old = r.Attributes()
		case types.Map:
			old = r.Elements()
		}
		values := map[string]attr.Value{}
		for k, v := range p.Elements() {
			mapped, err := honorPlanned(v, old[k])
			if err != nil {
				return nil, err
			}
			values[k] = mapped
		}
		value, d := types.MapValue(p.ElementType(ctx), values)
		if d.HasError() {
			return nil, fmt.Errorf("incompatible planned map")
		}
		return value, nil
	case types.List:
		values, err := honorElements(p.Elements(), remoteElements(remote))
		if err != nil {
			return nil, err
		}
		value, d := types.ListValue(p.ElementType(ctx), values)
		if d.HasError() {
			return nil, fmt.Errorf("incompatible planned list")
		}
		return value, nil
	case types.Tuple:
		values, err := honorElements(p.Elements(), remoteElements(remote))
		if err != nil {
			return nil, err
		}
		value, d := types.TupleValue(p.ElementTypes(ctx), values)
		if d.HasError() {
			return nil, fmt.Errorf("incompatible planned tuple")
		}
		return value, nil
	case types.Set:
		candidates := remoteElements(remote)
		used := map[int]bool{}
		values := make([]attr.Value, len(p.Elements()))
		for i, v := range p.Elements() {
			match := -1
			for j, c := range candidates {
				if !used[j] && knownMatch(v, c) {
					if match >= 0 {
						return nil, fmt.Errorf("ambiguous unknown set element")
					}
					match = j
				}
			}
			if match < 0 {
				return nil, fmt.Errorf("unmatched unknown set element")
			}
			used[match] = true
			mapped, err := honorPlanned(v, candidates[match])
			if err != nil {
				return nil, err
			}
			values[i] = mapped
		}
		value, d := types.SetValue(p.ElementType(ctx), values)
		if d.HasError() {
			return nil, fmt.Errorf("incompatible planned set")
		}
		return value, nil
	}
	return nil, fmt.Errorf("unsupported partially unknown planned value")
}
func remoteElements(v attr.Value) []attr.Value {
	switch x := v.(type) {
	case types.List:
		return x.Elements()
	case types.Tuple:
		return x.Elements()
	case types.Set:
		return x.Elements()
	}
	return nil
}
func honorElements(plan, remote []attr.Value) ([]attr.Value, error) {
	values := make([]attr.Value, len(plan))
	for i, p := range plan {
		var r attr.Value
		if i < len(remote) {
			r = remote[i]
		}
		value, err := honorPlanned(p, r)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}
func knownMatch(plan, remote attr.Value) bool {
	if plan.IsUnknown() {
		return remote != nil && plan.Type(context.Background()).Equal(remote.Type(context.Background()))
	}
	if p, ok := plan.(types.Object); ok {
		r, ok := remote.(types.Object)
		if !ok {
			return false
		}
		for k, v := range p.Attributes() {
			if !knownMatch(v, r.Attributes()[k]) {
				return false
			}
		}
		return true
	}
	if remote == nil {
		return false
	}
	return plan.Equal(remote)
}
