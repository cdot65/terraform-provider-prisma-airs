package provider

import (
	"context"
	"encoding/json"
	"github.com/cdot65/prisma-airs-go/aisec"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math/big"
	"strings"
	"testing"
)

func targetFixtureModel(family string, values map[string]attr.Value) RedTeamTargetResourceModel {
	model := RedTeamTargetResourceModel{Name: types.StringValue("fixture-target"), Description: types.StringValue(""), TargetType: types.StringNull(), APIEndpointType: types.StringValue("PUBLIC"), NetworkBrokerChannelUUID: types.StringNull()}
	for name, block := range model.blocks() {
		*block = types.ObjectNull(targetObjectTypes(name))
	}
	attrs := emptyTargetAttributes(family)
	for name, value := range values {
		attrs[name] = value
	}
	*model.blocks()[family] = types.ObjectValueMust(targetObjectTypes(family), attrs)
	return model
}
func TestTargetPayloadPreservesNativeShapeAndFalsyValues(t *testing.T) {
	nested := types.ObjectValueMust(map[string]attr.Type{"false": types.BoolType, "zero": types.NumberType, "null": types.StringType, "empty": types.TupleType{ElemTypes: []attr.Type{}}}, map[string]attr.Value{"false": types.BoolValue(false), "zero": types.NumberValue(big.NewFloat(0)), "null": types.StringNull(), "empty": types.TupleValueMust([]attr.Type{}, []attr.Value{})})
	payload := types.DynamicValue(nested)
	model := targetFixtureModel("rest", map[string]attr.Value{"api_endpoint": types.StringValue("https://example.com"), "request_body": payload, "response_body": types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"output": types.StringType}, map[string]attr.Value{"output": types.StringValue("{RESPONSE}")})), "response_key": types.StringValue("output")})
	var diagnostics diag.Diagnostics
	request := targetRequest(&model, &diagnostics)
	if diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	encoded, _ := request.ConnectionParams.Get()
	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	body := wire["request_json"].(map[string]any)
	if body["false"] != false || body["zero"] != float64(0) || body["null"] != nil || len(body["empty"].([]any)) != 0 {
		t.Error("falsy payload values were lost")
	}
	if !model.Rest.Attributes()["request_body"].Equal(payload) {
		t.Error("serialization changed concrete HCL state")
	}
	if family, _ := request.ConnectionType.Get(); family != "CUSTOM" {
		t.Error("REST transport must use the live CUSTOM discriminator")
	}
}
func TestTargetPayloadRejectsNestedUnknown(t *testing.T) {
	payload := types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"prompt": types.StringType}, map[string]attr.Value{"prompt": types.StringUnknown()}))
	model := targetFixtureModel("rest", map[string]attr.Value{"api_endpoint": types.StringValue("https://example.com"), "request_body": payload, "response_body": types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"output": types.StringType}, map[string]attr.Value{"output": types.StringValue("{RESPONSE}")})), "response_key": types.StringValue("output")})
	var diagnostics diag.Diagnostics
	targetRequest(&model, &diagnostics)
	if !diagnostics.HasError() {
		t.Error("unknown payload escaped to API")
	}
}
func TestTargetMaskedReadPreservesDesiredSecretsAndDetectsFixedDrift(t *testing.T) {
	model := targetFixtureModel("openai", map[string]attr.Value{"api_key": types.StringValue("desired-secret"), "model_name": types.StringValue("desired-model")})
	target := &rtschema.TargetRedact{UUID: "fixture-uuid", Name: "fixture-target", ConnectionType: aisec.Value(rtschema.TargetConnectionTypeOpenai), TargetType: aisec.Value(rtschema.TargetTypeModel), APIEndpointType: aisec.Value(rtschema.APIEndpointTypePublic), ResponseMode: aisec.Value(rtschema.ResponseModeRest), ConnectionParams: aisec.Value(rtschema.TargetRedactConnectionParams(`{"target_connection_config":{"api_key":"MASKED-BY-SERVICE","model_name":"external-model"}}`))}
	var diagnostics diag.Diagnostics
	mapTargetDetailsToState(context.Background(), target, &model, &diagnostics)
	if diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	attrs := model.OpenAI.Attributes()
	if attrs["api_key"].(types.String).ValueString() != "desired-secret" || attrs["model_name"].(types.String).ValueString() != "external-model" {
		t.Error("secret ownership or observable model drift failed")
	}
	imported := targetFixtureModel("openai", nil)
	mapTargetDetailsToState(context.Background(), target, &imported, &diagnostics)
	if !imported.OpenAI.Attributes()["api_key"].IsNull() {
		t.Error("import promoted masked secret to desired credential")
	}
}
func TestTargetDiagnosticsNeverEchoCredentials(t *testing.T) {
	err := aisec.NewHTTPError("credential=fixture-secret", aisec.ClientSideError, 400)
	if strings.Contains(safeTargetError(err), "fixture-secret") {
		t.Error("SDK response leaked a target credential")
	}
}
func TestTargetConnectionAndAuthenticationValidation(t *testing.T) {
	model := targetFixtureModel("rest", map[string]attr.Value{"api_endpoint": types.StringValue("https://example.com")})
	*model.blocks()["custom"] = model.Rest
	var diagnostics diag.Diagnostics
	validateTargetConfig(&model, &diagnostics)
	if !diagnostics.HasError() {
		t.Error("multiple families accepted")
	}
	model = targetFixtureModel("openai", map[string]attr.Value{"api_key": types.StringValue("secret"), "model_name": types.StringValue("model")})
	headers := emptyTargetAttributes("headers_auth")
	headers["headers"] = types.MapValueMust(types.StringType, map[string]attr.Value{"Authorization": types.StringValue("secret")})
	model.HeadersAuth = types.ObjectValueMust(targetObjectTypes("headers_auth"), headers)
	diagnostics = nil
	validateTargetConfig(&model, &diagnostics)
	if !diagnostics.HasError() {
		t.Error("native provider accepted a second auth mechanism")
	}
}

func TestTargetPayloadInvalidRootsAreRejectedDuringValidation(t *testing.T) {
	for _, payload := range []types.Dynamic{types.DynamicValue(types.StringValue("opaque-json")), types.DynamicValue(types.StringNull()), types.DynamicValue(types.TupleValueMust([]attr.Type{}, []attr.Value{}))} {
		model := targetFixtureModel("openai", map[string]attr.Value{"api_key": types.StringValue("secret"), "model_name": types.StringValue("model"), "request_body": payload})
		var diagnostics diag.Diagnostics
		validateTargetConfig(&model, &diagnostics)
		if !diagnostics.HasError() {
			t.Error("invalid root payload reached apply without a validation error")
		}
	}
}
