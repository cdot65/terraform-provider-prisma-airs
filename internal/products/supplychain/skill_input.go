package supplychain

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Convert native HCL, rejecting unknown nested values and JSON text before I/O.
func skillInput(v attr.Value) (any, error) {
	if v == nil || v.IsUnknown() {
		return nil, fmt.Errorf("registration_details must be fully known before applying")
	}
	if v.IsNull() {
		return nil, nil
	}
	switch x := v.(type) {
	case types.Dynamic:
		return skillInput(x.UnderlyingValue())
	case types.String:
		return x.ValueString(), nil
	case types.Bool:
		return x.ValueBool(), nil
	case types.Int64:
		return x.ValueInt64(), nil
	case types.Float64:
		return x.ValueFloat64(), nil
	case types.Number:
		return json.Number(x.ValueBigFloat().Text('f', -1)), nil
	case types.Object:
		return skillInputFields(x.Attributes())
	case types.Map:
		return skillInputFields(x.Elements())
	case types.List:
		return skillInputElements(x.Elements())
	case types.Tuple:
		return skillInputElements(x.Elements())
	case types.Set:
		return skillInputElements(x.Elements())
	}
	return nil, fmt.Errorf("unsupported native registration value")
}
func skillInputFields(items map[string]attr.Value) (map[string]any, error) {
	result := map[string]any{}
	for k, v := range items {
		item, e := skillInput(v)
		if e != nil {
			return nil, e
		}
		result[k] = item
	}
	return result, nil
}
func skillInputElements(items []attr.Value) ([]any, error) {
	result := make([]any, len(items))
	for i, v := range items {
		item, e := skillInput(v)
		if e != nil {
			return nil, e
		}
		result[i] = item
	}
	return result, nil
}

func skillHasUnknown(v attr.Value) bool {
	if v == nil || v.IsUnknown() {
		return true
	}
	if v.IsNull() {
		return false
	}
	switch x := v.(type) {
	case types.Dynamic:
		return skillHasUnknown(x.UnderlyingValue())
	case types.Object:
		for _, item := range x.Attributes() {
			if skillHasUnknown(item) {
				return true
			}
		}
	case types.Map:
		for _, item := range x.Elements() {
			if skillHasUnknown(item) {
				return true
			}
		}
	case types.Tuple:
		for _, item := range x.Elements() {
			if skillHasUnknown(item) {
				return true
			}
		}
	case types.List:
		for _, item := range x.Elements() {
			if skillHasUnknown(item) {
				return true
			}
		}
	case types.Set:
		for _, item := range x.Elements() {
			if skillHasUnknown(item) {
				return true
			}
		}
	}
	return false
}
