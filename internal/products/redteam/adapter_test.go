package redteam

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func adapterVars(fields map[string]struct {
	kind  string
	value types.String
}) types.Map {
	values := map[string]attr.Value{}
	for name, v := range fields {
		values[name] = types.ObjectValueMust(adapterVariableTypes, map[string]attr.Value{"type": types.StringValue(v.kind), "value": v.value})
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: adapterVariableTypes}, values)
}
func TestAdapterImportRetainsSecretInventoryAndPreservesStoredSecretsOnUpdate(t *testing.T) {
	redacted := true
	variables := []rtschema.AdapterVarResponse{{Key: "credential", Type: rtschema.AdapterVarTypeSecret, Value: aisec.Null[string](), IsRedacted: &redacted}, {Key: "endpoint", Type: rtschema.AdapterVarTypeVar, Value: aisec.Value("https://example.com")}}
	remote := &rtschema.CustomTargetAdapter{UUID: "adapter-fixture", Name: "fixture", ScriptB64: base64.StdEncoding.EncodeToString([]byte("print('fixture')\n")), Variables: &variables}
	var state adapterModel
	var d diag.Diagnostics
	mapAdapter(context.Background(), remote, &state, &d)
	if d.HasError() {
		t.Fatal(d)
	}
	if len(state.Variables.Elements()) != 2 || !state.Variables.Elements()["credential"].(types.Object).Attributes()["value"].IsNull() {
		t.Fatal("import lost secret key or invented value")
	}
	plan := state
	plan.Variables = adapterVars(map[string]struct {
		kind  string
		value types.String
	}{"credential": {"SECRET", types.StringNull()}, "endpoint": {"VAR", types.StringValue("https://updated.example.com")}})
	request := adapterRequest(&plan, &state, &d)
	if d.HasError() {
		t.Fatal(d)
	}
	for _, v := range *request.Variables {
		if v.Key == "credential" && !v.Value.IsNull() {
			t.Fatal("update must explicitly retain redacted secret")
		}
	}
	// An imported null cannot initialize a new secret or change its type.
	d = nil
	adapterRequest(&plan, nil, &d)
	if !d.HasError() {
		t.Fatal("incomplete create accepted")
	}
	d = nil
	plan.Variables = adapterVars(map[string]struct {
		kind  string
		value types.String
	}{"credential": {"VAR", types.StringNull()}})
	adapterRequest(&plan, &state, &d)
	if !d.HasError() {
		t.Fatal("null changed SECRET to VAR")
	}
}
func TestAdapterReadDetectsPlainVariableDriftAndKeepsConfiguredSecret(t *testing.T) {
	state := adapterModel{Variables: adapterVars(map[string]struct {
		kind  string
		value types.String
	}{"credential": {"SECRET", types.StringValue("original")}, "plain": {"VAR", types.StringValue("old")}, "removed": {"VAR", types.StringValue("old")}})}
	redacted := true
	values := []rtschema.AdapterVarResponse{{Key: "credential", Type: rtschema.AdapterVarTypeSecret, Value: aisec.Null[string](), IsRedacted: &redacted}, {Key: "plain", Type: rtschema.AdapterVarTypeVar, Value: aisec.Value("changed")}}
	var d diag.Diagnostics
	mapAdapterVariables(context.Background(), &rtschema.CustomTargetAdapter{Variables: &values}, &state, &d)
	if d.HasError() {
		t.Fatal(d)
	}
	if !state.Variables.Elements()["credential"].(types.Object).Attributes()["value"].Equal(types.StringValue("original")) || !state.Variables.Elements()["plain"].(types.Object).Attributes()["value"].Equal(types.StringValue("changed")) || len(state.Variables.Elements()) != 2 {
		t.Fatal("secret ownership or plain/key drift lost")
	}
}
func TestAdapterTargetRequestAndImport(t *testing.T) {
	model := targetFixtureModel("adapter", map[string]attr.Value{"uuid": types.StringValue("adapter-fixture")})
	model.APIEndpointType = types.StringValue("NETWORK_BROKER")
	model.NetworkBrokerChannelUUID = types.StringValue("channel-fixture")
	var d diag.Diagnostics
	req := targetRequest(&model, &d)
	if d.HasError() {
		t.Fatal(d)
	}
	kind, _ := req.ConnectionType.Get()
	uuid, _ := req.AdapterUUID.Get()
	if kind != rtschema.TargetConnectionTypeCustomTargetAdapter || uuid != "adapter-fixture" || !req.ConnectionParams.IsNull() || !req.ResponseMode.IsNull() {
		t.Fatal("adapter wire discriminators incorrect")
	}
	remote := &rtschema.TargetRedact{UUID: "target-fixture", Name: "fixture", ConnectionType: req.ConnectionType, APIEndpointType: req.APIEndpointType, NetworkBrokerChannelUUID: req.NetworkBrokerChannelUUID, AdapterUUID: req.AdapterUUID, ResponseMode: aisec.Value(rtschema.ResponseModeRest)}
	imported := targetFixtureModel("adapter", nil)
	mapTargetDetailsToState(context.Background(), remote, &imported, &d)
	if d.HasError() || !imported.Adapter.Attributes()["uuid"].Equal(types.StringValue("adapter-fixture")) {
		t.Fatal("adapter target import failed", d)
	}
}
