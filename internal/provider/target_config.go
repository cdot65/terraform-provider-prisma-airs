package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func selectedTargetBlock(model *RedTeamTargetResourceModel, names []string) (string, types.Object, bool) {
	var selected string
	var value types.Object
	unknown := false
	for _, name := range names {
		block := *model.blocks()[name]
		if block.IsUnknown() {
			unknown = true
		} else if !block.IsNull() {
			if selected != "" {
				return "multiple", value, unknown
			}
			selected, value = name, block
		}
	}
	return selected, value, unknown
}

func validateTargetConfig(model *RedTeamTargetResourceModel, diags *diag.Diagnostics) {
	family, block, unknown := selectedTargetBlock(model, targetFamilies)
	if family == "multiple" || (family == "" && !unknown) {
		diags.AddError("Choose one target connection", "Configure exactly one of openai, hugging_face, databricks, bedrock, custom, rest or streaming.")
	}
	auth, _, authUnknown := selectedTargetBlock(model, targetAuthBlocks)
	if auth == "multiple" {
		diags.AddError("Choose one authentication method", "Configure at most one of headers_auth, basic_auth or oauth2_auth.")
	}
	if family != "" && family != "multiple" && !unknown && !authUnknown && auth != "" && family != "custom" && family != "rest" && family != "streaming" {
		diags.AddError("Unsupported target authentication", "Native provider blocks contain their own credentials. Authentication blocks apply only to custom, rest and streaming connections.")
	}
	// Framework validates required children even when an optional block is
	// absent. Enforce conditional requirements on selected blocks here.
	if family != "" && family != "multiple" {
		required := map[string][]string{
			"openai": {"api_key", "model_name", "request_body", "response_body"}, "hugging_face": {"api_key", "model_name", "request_body", "response_body", "response_key"},
			"databricks": {"workspace_url", "model_name", "request_body", "response_body", "response_key", "response_stop_key", "response_stop_value"}, "bedrock": {"access_id", "access_secret", "region", "model_id", "request_body", "response_body", "response_key"},
			"rest": {"api_endpoint", "request_body", "response_body", "response_key"}, "custom": {"api_endpoint", "request_body", "response_body", "response_key"}, "streaming": {"api_endpoint", "request_body", "response_body", "response_key", "response_stop_key", "response_stop_value"},
		}[family]
		validateTargetRequired(family, block, required, diags)
	}
	if auth != "" && auth != "multiple" {
		_, authBlock, _ := selectedTargetBlock(model, targetAuthBlocks)
		required := map[string][]string{"headers_auth": {"headers"}, "basic_auth": {"username", "password"}, "oauth2_auth": {"token_url", "inject_header"}}[auth]
		validateTargetRequired(auth, authBlock, required, diags)
	}
	if family != "" && family != "multiple" && family != "custom" && family != "rest" && family != "streaming" && !model.TargetType.IsNull() && !model.TargetType.IsUnknown() && model.TargetType.ValueString() != "MODEL" {
		diags.AddError("Unsupported provider target category", "Provider connection blocks require target_type = MODEL. Omit target_type to infer it.")
	}
	endpoint := model.APIEndpointType
	channel := model.NetworkBrokerChannelUUID
	if !endpoint.IsUnknown() && !channel.IsUnknown() {
		if endpoint.ValueString() == "NETWORK_BROKER" && channel.IsNull() {
			diags.AddError("Network Broker channel required", "NETWORK_BROKER requires a preexisting network_broker_channel_uuid; this provider does not create permanent channels.")
		} else if !channel.IsNull() && endpoint.ValueString() != "NETWORK_BROKER" {
			diags.AddError("Unexpected Network Broker channel", "network_broker_channel_uuid is allowed only with api_endpoint_type = NETWORK_BROKER.")
		}
	}
	if family == "databricks" {
		attrs := block.Attributes()
		token, id, secret := attrs["access_token"], attrs["client_id"], attrs["secret"]
		if !token.IsUnknown() && !id.IsUnknown() && !secret.IsUnknown() {
			if (token.IsNull() && (id.IsNull() || secret.IsNull())) || (!token.IsNull() && (!id.IsNull() || !secret.IsNull())) {
				diags.AddError("Invalid Databricks credentials", "Supply access_token, or both client_id and secret for OAuth.")
			}
		}
	}
	for name, candidate := range model.blocks() {
		if candidate.IsNull() || candidate.IsUnknown() {
			continue
		}
		for field, value := range candidate.Attributes() {
			payload, ok := value.(types.Dynamic)
			if !ok || payload.IsNull() || payload.IsUnknown() || payload.IsUnderlyingValueUnknown() {
				continue
			}
			if payload.IsUnderlyingValueNull() {
				diags.AddError("Invalid target payload", name+"."+field+" must be a native HCL object or map, not null.")
				continue
			}
			switch payload.UnderlyingValue().(type) {
			case types.Object, types.Map:
			default:
				diags.AddError("Invalid target payload", name+"."+field+" must be a native HCL object or map; nested lists are supported.")
			}
		}
	}
}

// Convert values recursively without coercing their HCL type in Terraform state.
// Unknowns must resolve before apply; explicit nulls remain JSON null, and false,
// zero, empty objects and empty lists are never collapsed into omission.
func targetJSONValue(value attr.Value) (any, error) {
	if value.IsUnknown() {
		return nil, fmt.Errorf("value is unknown")
	}
	if value.IsNull() {
		return nil, nil
	}
	switch value := value.(type) {
	case types.Dynamic:
		return targetJSONValue(value.UnderlyingValue())
	case types.String:
		return value.ValueString(), nil
	case types.Bool:
		return value.ValueBool(), nil
	case types.Int64:
		return value.ValueInt64(), nil
	case types.Float64:
		return value.ValueFloat64(), nil
	case types.Number:
		return json.Number(value.ValueBigFloat().Text('g', -1)), nil
	case types.Object:
		return targetJSONAttributes(value.Attributes())
	case types.Map:
		return targetJSONAttributes(value.Elements())
	case types.Tuple:
		return targetJSONElements(value.Elements())
	case types.List:
		return targetJSONElements(value.Elements())
	case types.Set:
		return targetJSONElements(value.Elements())
	default:
		return nil, fmt.Errorf("unsupported HCL value type")
	}
}

func targetJSONAttributes(attrs map[string]attr.Value) (map[string]any, error) {
	result := map[string]any{}
	for name, value := range attrs {
		encoded, err := targetJSONValue(value)
		if err != nil {
			return nil, err
		}
		result[name] = encoded
	}
	return result, nil
}

func targetJSONElements(elements []attr.Value) ([]any, error) {
	result := make([]any, len(elements))
	for i, value := range elements {
		encoded, err := targetJSONValue(value)
		if err != nil {
			return nil, err
		}
		result[i] = encoded
	}
	return result, nil
}

func targetObjectJSON(block types.Object, names map[string]string, diags *diag.Diagnostics) map[string]any {
	result := map[string]any{}
	for name, value := range block.Attributes() {
		// Optional top-level fields are omitted; nested nulls are preserved.
		if value.IsNull() {
			continue
		}
		encoded, err := targetJSONValue(value)
		if err != nil {
			diags.AddError("Unresolved target configuration", "All connection settings must be known before apply.")
			continue
		}
		if dynamic, ok := value.(types.Dynamic); ok {
			if _, ok := encoded.(map[string]any); !ok && !dynamic.IsUnderlyingValueNull() {
				diags.AddError("Invalid target payload", name+" must be a native HCL object or map; nested lists are supported.")
				continue
			}
		}
		if wireName, ok := names[name]; ok {
			name = wireName
		}
		result[name] = encoded
	}
	return result
}

func targetRequest(model *RedTeamTargetResourceModel, diags *diag.Diagnostics) rtschema.TargetCreateRequest {
	validateTargetConfig(model, diags)
	family, block, unknown := selectedTargetBlock(model, targetFamilies)
	if unknown || family == "" || family == "multiple" || diags.HasError() {
		if !diags.HasError() {
			diags.AddError("Unresolved target connection", "Choose one fully known connection before apply.")
		}
		return rtschema.TargetCreateRequest{}
	}
	params := targetObjectJSON(block, map[string]string{"request_body": "request_json", "response_body": "response_json"}, diags)
	response := "REST"
	if family == "streaming" || family == "databricks" {
		response = "STREAMING"
	}
	if family == "databricks" {
		params["auth_type"] = "ACCESS_TOKEN"
		if block.Attributes()["access_token"].IsNull() {
			params["auth_type"] = "OAUTH"
		}
	}
	if family != "rest" && family != "custom" && family != "streaming" {
		provider := map[string]any{}
		for _, key := range []string{"api_key", "model_name", "auth_type", "access_token", "client_id", "secret", "workspace_url", "access_id", "access_secret", "session_token", "region", "model_id"} {
			if value, exists := params[key]; exists {
				provider[key] = value
				delete(params, key)
			}
		}
		params["target_connection_config"] = provider
	}
	wireFamily := strings.ToUpper(family)
	if family == "rest" || family == "streaming" {
		wireFamily = "CUSTOM"
	}
	targetType := model.TargetType.ValueString()
	if model.TargetType.IsNull() || model.TargetType.IsUnknown() {
		targetType = "APPLICATION"
		if family != "rest" && family != "custom" && family != "streaming" {
			targetType = "MODEL"
		}
	}

	encoded, err := json.Marshal(params)
	if err != nil {
		diags.AddError("Invalid target payload", "Unable to serialize the desired connection.")
	}
	body := rtschema.TargetCreateRequest{
		Name: model.Name.ValueString(), Description: aisec.Value(model.Description.ValueString()),
		TargetType:               aisec.Value(rtschema.TargetType(targetType)),
		ConnectionType:           aisec.Value(rtschema.TargetConnectionType(wireFamily)),
		APIEndpointType:          aisec.Value(rtschema.APIEndpointType(model.APIEndpointType.ValueString())),
		ResponseMode:             aisec.Value(rtschema.ResponseMode(response)),
		ConnectionParams:         aisec.Value(rtschema.TargetCreateRequestConnectionParams(encoded)),
		NetworkBrokerChannelUUID: aisec.Null[string](),
		AuthType:                 aisec.Null[rtschema.RedTeamSharedSchemasAuthConfigAuthType](),
		AuthConfig:               aisec.Null[rtschema.TargetCreateRequestAuthConfig](),
	}
	if !model.NetworkBrokerChannelUUID.IsNull() {
		body.NetworkBrokerChannelUUID = aisec.Value(model.NetworkBrokerChannelUUID.ValueString())
	}
	auth, authBlock, authUnknown := selectedTargetBlock(model, targetAuthBlocks)
	if authUnknown {
		diags.AddError("Unresolved target authentication", "Authentication must be known before apply.")
	}
	if auth != "" && auth != "multiple" {
		var mapping map[string]string
		var kind string
		switch auth {
		case "headers_auth":
			kind, mapping = "HEADERS", map[string]string{"headers": "auth_header"}
		case "basic_auth":
			kind, mapping = "BASIC_AUTH", map[string]string{"username": "basic_username", "password": "basic_password", "location": "basic_auth_location"}
		case "oauth2_auth":
			kind, mapping = "OAUTH2", map[string]string{"token_url": "oauth2_token_url", "headers": "oauth2_headers", "body": "oauth2_body_params", "expiry_minutes": "oauth2_expiry_minutes", "response_key": "oauth2_token_response_key", "inject_header": "oauth2_inject_header"}
		}
		encoded, err := json.Marshal(targetObjectJSON(authBlock, mapping, diags))
		if err != nil {
			diags.AddError("Invalid target authentication", "Unable to serialize desired authentication settings.")
		}
		body.AuthType = aisec.Value(rtschema.RedTeamSharedSchemasAuthConfigAuthType(kind))
		body.AuthConfig = aisec.Value(rtschema.TargetCreateRequestAuthConfig(encoded))
	}
	return body
}

func validateTargetRequired(name string, block types.Object, required []string, diags *diag.Diagnostics) {
	for _, field := range required {
		if value := block.Attributes()[field]; value.IsNull() {
			diags.AddError("Missing target setting", name+"."+field+" is required when this block is configured. Imported secrets must be supplied explicitly.")
		}
	}
}
