package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// nativeJSON keeps JSON at the SDK seam. HCL objects, tuples, numbers and nulls
// remain native values in configuration, plans and state.
func nativeJSON(v attr.Value) (any, error) {
	if v.IsUnknown() {
		return nil, fmt.Errorf("a Gateway input is still unknown")
	}
	if v.IsNull() {
		return nil, nil
	}
	switch x := v.(type) {
	case types.Dynamic:
		return nativeJSON(x.UnderlyingValue())
	case types.String:
		return x.ValueString(), nil
	case types.Bool:
		return x.ValueBool(), nil
	case types.Int64:
		return x.ValueInt64(), nil
	case types.Float64:
		return x.ValueFloat64(), nil
	case types.Number:
		return json.Number(x.ValueBigFloat().Text('g', -1)), nil
	case types.Object:
		return nativeAttributes(x.Attributes())
	case types.Map:
		return nativeAttributes(x.Elements())
	case types.List:
		return nativeElements(x.Elements())
	case types.Set:
		return nativeElements(x.Elements())
	case types.Tuple:
		return nativeElements(x.Elements())
	default:
		return nil, fmt.Errorf("unsupported Gateway HCL type")
	}
}

func nativeAttributes(values map[string]attr.Value) (map[string]any, error) {
	result := map[string]any{}
	for k, v := range values {
		x, err := nativeJSON(v)
		if err != nil {
			return nil, err
		}
		result[k] = x
	}
	return result, nil
}
func nativeElements(values []attr.Value) ([]any, error) {
	result := make([]any, len(values))
	for i, v := range values {
		x, err := nativeJSON(v)
		if err != nil {
			return nil, err
		}
		result[i] = x
	}
	return result, nil
}
func nullNative(ctx context.Context, t attr.Type) (attr.Value, error) {
	return t.ValueFromTerraform(ctx, tftypes.NewValue(t.TerraformType(ctx), nil))
}

// readNative uses the prior concrete HCL type wherever possible. This avoids
// converting an empty tuple into a list or an integer into a string on refresh.
// New remote keys and removed keys are retained as drift, never projected away.
func readNative(ctx context.Context, value any, prior attr.Value) (attr.Value, error) {
	if dynamic, ok := prior.(types.Dynamic); ok {
		var template attr.Value
		if !dynamic.IsNull() && !dynamic.IsUnknown() {
			template = dynamic.UnderlyingValue()
		}
		v, err := readNative(ctx, value, template)
		if err != nil {
			return nil, err
		}
		return types.DynamicValue(v), nil
	}
	if value == nil {
		if prior != nil {
			return nullNative(ctx, prior.Type(ctx))
		}
		return types.StringNull(), nil
	}
	switch x := value.(type) {
	case string:
		return types.StringValue(x), nil
	case bool:
		return types.BoolValue(x), nil
	case json.Number:
		if _, ok := prior.(types.Int64); ok {
			n, err := x.Int64()
			return types.Int64Value(n), err
		}
		if _, ok := prior.(types.Float64); ok {
			n, err := x.Float64()
			return types.Float64Value(n), err
		}
		n, _, err := big.ParseFloat(string(x), 10, 512, big.ToNearestEven)
		if err != nil {
			return nil, err
		}
		return types.NumberValue(n), nil
	case float64:
		return readNative(ctx, json.Number(fmt.Sprint(x)), prior)
	case map[string]any:
		old := map[string]attr.Value{}
		switch p := prior.(type) {
		case types.Object:
			old = p.Attributes()
		case types.Map:
			old = p.Elements()
		}
		values := map[string]attr.Value{}
		ts := map[string]attr.Type{}
		for k, item := range x {
			v, err := readNative(ctx, item, old[k])
			if err != nil {
				return nil, err
			}
			values[k] = v
			ts[k] = v.Type(ctx)
		}
		if p, ok := prior.(types.Map); ok {
			v, d := types.MapValue(p.ElementType(ctx), values)
			if d.HasError() {
				return nil, fmt.Errorf("gateway map contains incompatible remote values")
			}
			return v, nil
		}
		v, d := types.ObjectValue(ts, values)
		if d.HasError() {
			return nil, fmt.Errorf("cannot read Gateway object")
		}
		return v, nil
	case []any:
		old := []attr.Value{}
		switch p := prior.(type) {
		case types.Tuple:
			old = p.Elements()
		case types.List:
			old = p.Elements()
		case types.Set:
			old = p.Elements()
		}
		values := make([]attr.Value, len(x))
		ts := make([]attr.Type, len(x))
		for i, item := range x {
			var p attr.Value
			if i < len(old) {
				p = old[i]
			}
			v, err := readNative(ctx, item, p)
			if err != nil {
				return nil, err
			}
			values[i] = v
			ts[i] = v.Type(ctx)
		}
		switch p := prior.(type) {
		case types.List:
			v, d := types.ListValue(p.ElementType(ctx), values)
			if d.HasError() {
				return nil, fmt.Errorf("cannot read Gateway list")
			}
			return v, nil
		case types.Set:
			v, d := types.SetValue(p.ElementType(ctx), values)
			if d.HasError() {
				return nil, fmt.Errorf("cannot read Gateway set")
			}
			return v, nil
		}
		v, d := types.TupleValue(ts, values)
		if d.HasError() {
			return nil, fmt.Errorf("cannot read Gateway tuple")
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported Gateway response value")
	}
}

// Routing documents are visible in plans. Never include the offending value
// in a diagnostic: validate keys before a write or before importing a document.
func routingCredentials(value any) bool {
	switch x := value.(type) {
	case map[string]any:
		for k, v := range x {
			key := strings.ReplaceAll(strings.ToLower(k), "-", "_")
			if v != nil && credentialField(key) {
				return true
			}
			if routingCredentials(v) {
				return true
			}
		}
	case []any:
		for _, v := range x {
			if routingCredentials(v) {
				return true
			}
		}
	}
	return false
}

func credentialField(key string) bool {
	// Identifiers point to separately managed credentials; they are safe to show.
	if key == "virtual_key" || key == "provider" || key == "secret_reference_id" || strings.HasSuffix(key, "_secret_reference_id") {
		return false
	}
	compact := strings.ReplaceAll(key, "_", "")
	for _, suffix := range []string{"apikey", "apitoken", "token", "secret", "password", "privatekey", "secretkey", "secretaccesskey", "credentials", "authorization"} {
		if strings.HasSuffix(compact, suffix) {
			return true
		}
	}
	return strings.Contains(compact, "serviceaccount") || compact == "clientsecret"
}
