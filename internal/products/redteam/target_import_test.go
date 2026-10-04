package redteam

import (
	"context"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTargetImportRecoversObservablePayloadsAndAllowsReadOnlyConfig(t *testing.T) {
	model := targetFixtureModel("openai", nil)
	target := &rtschema.TargetRedact{UUID: "fixture", Name: "fixture", ConnectionType: aisec.Value(rtschema.TargetConnectionTypeOpenai), ResponseMode: aisec.Value(rtschema.ResponseModeRest), ConnectionParams: aisec.Value(rtschema.TargetRedactConnectionParams(`{"target_connection_config":{"api_key":"********","model_name":"model"},"request_json":{"prompt":"{INPUT}","nullable":null},"response_json":{"output":"{RESPONSE}"},"response_key":"output"}`))}
	var d diag.Diagnostics
	mapTargetDetailsToState(context.Background(), target, &model, &d)
	if d.HasError() {
		t.Fatal(d)
	}
	if model.OpenAI.Attributes()["request_body"].IsNull() || model.OpenAI.Attributes()["response_body"].IsNull() {
		t.Fatal("observable payloads lost during import")
	}
	if !model.OpenAI.Attributes()["api_key"].IsNull() {
		t.Fatal("masked secret imported")
	}
	validateTargetConfig(&model, &d)
	if d.HasError() {
		t.Fatal("read-only imported config must validate", d)
	}
	targetRequest(&model, &d)
	if !d.HasError() {
		t.Fatal("incomplete write was allowed")
	}
}

func TestTargetRedactedHeadersBlockIncompleteMetadataUpdates(t *testing.T) {
	model := targetFixtureModel("rest", nil)
	remote := &rtschema.TargetRedact{UUID: "fixture", Name: "fixture", ConnectionType: aisec.Value(rtschema.TargetConnectionTypeCustom), ResponseMode: aisec.Value(rtschema.ResponseModeRest), ConnectionParams: aisec.Value(rtschema.TargetRedactConnectionParams(`{"api_endpoint":"https://example.com","request_json":{"prompt":"{INPUT}"},"response_json":{"output":"{RESPONSE}"},"response_key":"output","request_headers":{"Authorization":"********"}}`))}
	var d diag.Diagnostics
	mapTargetDetailsToState(context.Background(), remote, &model, &d)
	if d.HasError() {
		t.Fatal(d)
	}
	if len(model.UnavailableFields.Elements()) != 1 || !model.Rest.Attributes()["request_headers"].IsNull() {
		t.Fatal("redacted header observation was lost or imported as a credential")
	}
	plan := model
	plan.Description = types.StringValue("edited metadata")
	validateTargetUnavailable(&model, &plan, &d)
	if !d.HasError() {
		t.Fatal("metadata update could omit the redacted header")
	}
	attrs := plan.Rest.Attributes()
	attrs["request_headers"] = types.MapValueMust(types.StringType, map[string]attr.Value{"Authorization": types.StringValue("original-value")})
	plan.Rest = types.ObjectValueMust(targetObjectTypes("rest"), attrs)
	d = nil
	validateTargetUnavailable(&model, &plan, &d)
	targetRequest(&plan, &d)
	if d.HasError() {
		t.Fatal("complete inputs rejected", d)
	}
}

func TestImportedEmptyStreamingStopFieldsAllowObservationButBlockWrites(t *testing.T) {
	payload := types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}))
	model := targetFixtureModel("streaming", map[string]attr.Value{"api_endpoint": types.StringValue("https://example.com"), "request_body": payload, "response_body": payload, "response_key": types.StringValue("output"), "response_stop_key": types.StringValue(""), "response_stop_value": types.StringValue("")})
	var d diag.Diagnostics
	validateTargetConfig(&model, &d)
	if d.HasError() {
		t.Fatal("observed stop fields rejected", d)
	}
	targetRequest(&model, &d)
	if !d.HasError() {
		t.Fatal("empty stop fields reached write path")
	}
}
