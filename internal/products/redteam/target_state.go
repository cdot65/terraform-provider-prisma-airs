package redteam

import (
	"context"
	"encoding/json"
	"strings"

	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func emptyTargetAttributes(name string) map[string]attr.Value {
	values := map[string]attr.Value{}
	for key, kind := range targetObjectTypes(name) {
		switch kind := kind.(type) {
		case basetypes.StringType:
			values[key] = types.StringNull()
		case basetypes.Int64Type:
			values[key] = types.Int64Null()
		case types.MapType:
			values[key] = types.MapNull(kind.ElemType)
		case basetypes.DynamicType:
			values[key] = types.DynamicNull()
		}
	}
	return values
}

// Reconcile fixed fields and recover usable payloads on import. Configured
// payloads retain their concrete HCL types, and masks never become credentials.
func mapTargetDetailsToState(ctx context.Context, target *rtschema.TargetRedact, state *RedTeamTargetResourceModel, diags *diag.Diagnostics) {
	state.UnavailableFields = types.ListValueMust(types.StringType, []attr.Value{})
	state.ID = types.StringValue(target.UUID)
	state.UUID = state.ID
	state.Name = types.StringValue(target.Name)
	description, _ := target.Description.Get()
	state.Description = types.StringValue(description)
	if value, ok := target.TargetType.Get(); ok {
		state.TargetType = types.StringValue(string(value))
	}
	if value, ok := target.APIEndpointType.Get(); ok {
		state.APIEndpointType = types.StringValue(string(value))
	}
	state.ResponseMode = types.StringNull()
	if value, ok := target.ResponseMode.Get(); ok {
		state.ResponseMode = types.StringValue(string(value))
	}
	if value, ok := target.NetworkBrokerChannelUUID.Get(); ok {
		state.NetworkBrokerChannelUUID = types.StringValue(value)
	} else {
		state.NetworkBrokerChannelUUID = types.StringNull()
	}
	state.Status = types.StringValue(string(target.Status))
	state.CreatedAt = types.StringValue(target.CreatedAt)
	state.UpdatedAt = types.StringValue(target.UpdatedAt)
	connection, ok := target.ConnectionType.Get()
	if !ok {
		diags.AddError("Missing target connection type", "The service did not identify the connection family.")
		return
	}
	state.ConnectionType = types.StringValue(string(connection))
	family := strings.ToLower(string(connection))
	if connection == rtschema.TargetConnectionTypeCustomTargetAdapter {
		family = "adapter"
	}
	selected, _, _ := selectedTargetBlock(state, targetFamilies)
	mode, _ := target.ResponseMode.Get()
	if family == "custom" {
		if mode == rtschema.ResponseModeStreaming {
			family = "streaming"
		} else if selected == "rest" {
			family = "rest"
		}
	}
	block, supported := state.blocks()[family]
	if !supported {
		diags.AddError("Unsupported target connection type", "This provider supports openai, hugging_face, databricks, bedrock, custom, rest, streaming and adapter.")
		return
	}
	for _, name := range targetFamilies {
		if name != family {
			*state.blocks()[name] = types.ObjectNull(targetObjectTypes(name))
		}
	}
	attrs := emptyTargetAttributes(family)
	if !block.IsNull() && !block.IsUnknown() {
		attrs = block.Attributes()
	}
	if family == "adapter" {
		if uuid, ok := target.AdapterUUID.Get(); ok {
			attrs["uuid"] = types.StringValue(uuid)
		}
		if overrides, ok := target.AdapterVariableOverrides.Get(); ok && (len(overrides) != 0 || !attrs["variable_overrides"].IsNull()) {
			synthetic := &rtschema.CustomTargetAdapter{Variables: &overrides}
			prior := adapterModel{Variables: attrs["variable_overrides"].(types.Map)}
			mapAdapterVariables(ctx, synthetic, &prior, diags)
			attrs["variable_overrides"] = prior.Variables
		}
		var d diag.Diagnostics
		*block, d = types.ObjectValue(targetObjectTypes(family), attrs)
		diags.Append(d...)
		return
	}
	var remote map[string]json.RawMessage
	if value, ok := target.ConnectionParams.Get(); ok {
		if err := json.Unmarshal(value, &remote); err != nil {
			diags.AddError("Invalid target read response", "The service returned invalid connection settings.")
			return
		}
	}
	if family != "rest" && family != "custom" && family != "streaming" {
		if config, ok := remote["target_connection_config"]; ok {
			var provider map[string]json.RawMessage
			if err := json.Unmarshal(config, &provider); err != nil {
				diags.AddError("Invalid provider read response", "The service returned invalid native provider settings.")
				return
			}
			for key, value := range provider {
				remote[key] = value
			}
		}
	}
	fixed := []string{"model_name", "workspace_url", "region", "model_id", "api_endpoint", "response_key", "response_stop_key", "response_stop_value"}
	readTargetStrings(attrs, remote, fixed, nil)
	readTargetObservable(ctx, attrs, remote, map[string]string{"request_body": "request_json", "response_body": "response_json", "request_headers": "request_headers"}, diags)
	var d diag.Diagnostics
	*block, d = types.ObjectValue(targetObjectTypes(family), attrs)
	diags.Append(d...)
	mapTargetAuthToState(ctx, target, state, diags)
	var authRemote map[string]json.RawMessage
	if raw, ok := target.AuthConfig.Get(); ok {
		if json.Unmarshal(raw, &authRemote) != nil {
			diags.AddError("Invalid target authentication", "The service returned invalid authentication settings.")
		}
	}
	authName, _, _ := selectedTargetBlock(state, targetAuthBlocks)
	state.UnavailableFields = targetUnavailableFields(ctx, family, remote, authName, authRemote, diags)
}

func readTargetStrings(attrs map[string]attr.Value, remote map[string]json.RawMessage, fixed []string, mapping map[string]string) {
	for _, name := range fixed {
		if _, exists := attrs[name]; !exists {
			continue
		}
		key := name
		if mapped, ok := mapping[name]; ok {
			key = mapped
		}
		if raw, present := remote[key]; present {
			var value *string
			if json.Unmarshal(raw, &value) == nil {
				if value == nil {
					attrs[name] = types.StringNull()
				} else {
					attrs[name] = types.StringValue(*value)
				}
			}
		}
	}
}

func mapTargetAuthToState(ctx context.Context, target *rtschema.TargetRedact, state *RedTeamTargetResourceModel, diags *diag.Diagnostics) {
	kind, _ := target.AuthType.Get()
	name := map[string]string{"HEADERS": "headers_auth", "BASIC_AUTH": "basic_auth", "OAUTH2": "oauth2_auth"}[string(kind)]
	for _, other := range targetAuthBlocks {
		if other != name {
			*state.blocks()[other] = types.ObjectNull(targetObjectTypes(other))
		}
	}
	if name == "" {
		return
	}
	block := state.blocks()[name]
	attrs := emptyTargetAttributes(name)
	if !block.IsNull() && !block.IsUnknown() {
		attrs = block.Attributes()
	}
	var remote map[string]json.RawMessage
	if value, ok := target.AuthConfig.Get(); ok {
		if err := json.Unmarshal(value, &remote); err != nil {
			diags.AddError("Invalid authentication read response", "The service returned invalid authentication settings.")
			return
		}
	}
	readTargetStrings(attrs, remote, []string{"token_url", "response_key", "location"}, map[string]string{"token_url": "oauth2_token_url", "response_key": "oauth2_token_response_key", "location": "basic_auth_location"})
	if name == "oauth2_auth" {
		readTargetObservable(ctx, attrs, remote, map[string]string{"inject_header": "oauth2_inject_header", "headers": "oauth2_headers", "body": "oauth2_body_params"}, diags)
		if raw, ok := remote["oauth2_expiry_minutes"]; ok {
			var value *int64
			if json.Unmarshal(raw, &value) == nil {
				attrs["expiry_minutes"] = types.Int64Null()
				if value != nil {
					attrs["expiry_minutes"] = types.Int64Value(*value)
				}
			}
		}
	}
	var d diag.Diagnostics
	*block, d = types.ObjectValue(targetObjectTypes(name), attrs)
	diags.Append(d...)
}
