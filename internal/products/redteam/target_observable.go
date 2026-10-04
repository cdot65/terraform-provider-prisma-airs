package redteam

import (
	"context"
	"encoding/json"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Masked values are observation markers, never usable credentials.
func targetContainsMask(value attr.Value) bool {
	if value == nil || value.IsNull() || value.IsUnknown() {
		return false
	}
	// Use the write serializer so every supported HCL collection is inspected.
	decoded, err := targetJSONValue(value)
	return err == nil && len(targetMaskedPaths(decoded, "")) != 0
}

// JSON Pointer suffixes retain masked leaf identities, including escaped keys.
func targetMaskedPaths(value any, pointer string) []string {
	var paths []string
	switch v := value.(type) {
	case string:
		text := strings.TrimSpace(v)
		upper := strings.ToUpper(text)
		if strings.Contains(text, "***") || upper == "MASKED-BY-SERVICE" || upper == "<REDACTED>" || upper == "[REDACTED]" {
			paths = append(paths, pointer)
		}
	case map[string]any:
		for key, item := range v {
			escaped := strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
			paths = append(paths, targetMaskedPaths(item, pointer+"/"+escaped)...)
		}
	case []any:
		for i, item := range v {
			paths = append(paths, targetMaskedPaths(item, pointer+"/"+strconv.Itoa(i))...)
		}
	}
	return paths
}

// Recover usable payloads on import. Configured payloads retain their concrete
// HCL shape because masked/normalized reads cannot establish payload drift.
func readTargetObservable(ctx context.Context, attrs map[string]attr.Value, remote map[string]json.RawMessage, mapping map[string]string, diags *diag.Diagnostics) {
	for name, key := range mapping {
		prior, exists := attrs[name]
		if !exists || !prior.IsNull() {
			continue
		}
		raw, exists := remote[key]
		if !exists || string(raw) == "null" {
			continue
		}
		var value any
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			diags.AddError("Invalid target read response", "The service returned an invalid observable payload.")
			continue
		}
		mapped := targetReadValue(ctx, value, diags)
		if diags.HasError() {
			continue
		}
		if _, ok := prior.(types.Dynamic); ok {
			mapped = types.DynamicValue(mapped)
		}
		if _, ok := prior.(types.Map); ok {
			fields, ok := value.(map[string]any)
			if !ok {
				continue
			}
			items := map[string]attr.Value{}
			for k, v := range fields {
				if text, ok := v.(string); ok {
					items[k] = types.StringValue(text)
				} else {
					diags.AddError("Invalid target headers", "The service returned a nonstring header value.")
				}
			}
			mapped = types.MapValueMust(types.StringType, items)
		}
		if !targetContainsMask(mapped) {
			attrs[name] = mapped
		}
	}
}

func targetReadValue(ctx context.Context, value any, diags *diag.Diagnostics) attr.Value {
	switch v := value.(type) {
	case nil:
		return types.DynamicNull()
	case string:
		return types.StringValue(v)
	case bool:
		return types.BoolValue(v)
	case json.Number:
		n, _, err := big.ParseFloat(string(v), 10, 512, big.ToNearestEven)
		if err != nil {
			diags.AddError("Invalid target number", "The service returned an invalid number.")
			return types.NumberNull()
		}
		return types.NumberValue(n)
	case map[string]any:
		kinds := map[string]attr.Type{}
		values := map[string]attr.Value{}
		for k, item := range v {
			values[k] = targetReadValue(ctx, item, diags)
			kinds[k] = values[k].Type(ctx)
		}
		return types.ObjectValueMust(kinds, values)
	case []any:
		kinds := make([]attr.Type, len(v))
		values := make([]attr.Value, len(v))
		for i, item := range v {
			values[i] = targetReadValue(ctx, item, diags)
			kinds[i] = values[i].Type(ctx)
		}
		return types.TupleValueMust(kinds, values)
	default:
		diags.AddError("Invalid target value", "The service returned an unsupported value.")
		return types.DynamicNull()
	}
}

// Remember redacted payload/header paths so a later metadata update cannot
// silently omit an unavailable credential from a full replacement request.
func targetUnavailableFields(ctx context.Context, family string, params map[string]json.RawMessage, authName string, auth map[string]json.RawMessage, diags *diag.Diagnostics) types.List {
	paths := []string{}
	collect := func(prefix string, remote map[string]json.RawMessage, mapping map[string]string) {
		for field, key := range mapping {
			raw, exists := remote[key]
			if !exists || string(raw) == "null" {
				continue
			}
			var value any
			decoder := json.NewDecoder(strings.NewReader(string(raw)))
			decoder.UseNumber()
			if decoder.Decode(&value) != nil {
				continue
			}
			for _, pointer := range targetMaskedPaths(value, "") {
				paths = append(paths, prefix+"."+field+pointer)
			}
		}
	}
	collect(family, params, map[string]string{"request_body": "request_json", "response_body": "response_json", "request_headers": "request_headers", "api_key": "api_key", "access_id": "access_id", "access_secret": "access_secret", "session_token": "session_token", "access_token": "access_token", "client_id": "client_id", "secret": "secret"})
	collect(authName, auth, map[string]string{"headers": "auth_header", "username": "basic_username", "password": "basic_password", "body": "oauth2_body_params", "inject_header": "oauth2_inject_header"})
	if authName == "oauth2_auth" {
		collect(authName, auth, map[string]string{"headers": "oauth2_headers"})
	}
	sort.Strings(paths)
	value, d := types.ListValueFrom(ctx, types.StringType, paths)
	diags.Append(d...)
	return value
}

func validateTargetUnavailable(state, plan *RedTeamTargetResourceModel, diags *diag.Diagnostics) {
	for _, item := range state.UnavailableFields.Elements() {
		recorded := item.(types.String).ValueString()
		field, pointer, hasPointer := strings.Cut(recorded, "/")
		parts := strings.SplitN(field, ".", 2)
		if len(parts) != 2 {
			continue
		}
		blockName := parts[0]
		family, _, _ := selectedTargetBlock(plan, targetFamilies)
		if endpointTargetFamily(blockName) && endpointTargetFamily(family) {
			// REST, custom, and streaming share CUSTOM ownership in the API.
			blockName = family
		}
		block, exists := plan.blocks()[blockName]
		complete := false
		if exists && !block.IsNull() && !block.IsUnknown() {
			if value, exists := block.Attributes()[parts[1]]; exists {
				decoded, err := targetJSONValue(value)
				if err == nil {
					if hasPointer {
						decoded = targetPointerValue(decoded, strings.Split(pointer, "/"))
					}
					text, ok := decoded.(string)
					complete = ok && strings.TrimSpace(text) != "" && len(targetMaskedPaths(text, "")) == 0
				}
			}
		}
		if !complete {
			diags.AddError("Unavailable imported target input", recorded+" was redacted by the service. Supply its original complete value before updating this target.")
		}
	}
}

func targetPointerValue(value any, segments []string) any {
	for _, segment := range segments {
		switch v := value.(type) {
		case map[string]any:
			key := strings.ReplaceAll(strings.ReplaceAll(segment, "~1", "/"), "~0", "~")
			value = v[key]
		case []any:
			i, err := strconv.Atoi(segment)
			if err != nil || i < 0 || i >= len(v) {
				return nil
			}
			value = v[i]
		default:
			return nil
		}
	}
	return value
}
