package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func configFixture(t *testing.T, handler http.HandlerFunc) (*gatewayResource, resource.SchemaResponse) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			return
		}
		if r.Header.Get("x-tsg-id") != "123" {
			t.Error("missing shared tenant header")
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	sdk, e := gw.NewClient(gw.Opts{ClientID: "client", ClientSecret: "secret", TsgID: "123", DataEndpoint: server.URL, AdminEndpoint: server.URL, TokenEndpoint: server.URL + "/token", NumRetries: 1})
	if e != nil {
		t.Fatal(e)
	}
	r := &gatewayResource{definition: definitions()[0], client: &client{sdk: sdk, organisation: "123"}}
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	return r, s
}
func modelFixture(t *testing.T, r *gatewayResource, values map[string]attr.Value) types.Object {
	t.Helper()
	ctx := context.Background()
	ts := map[string]attr.Type{}
	vs := map[string]attr.Value{}
	for k, a := range r.attributes() {
		ts[k] = a.GetType()
		v, e := nullNative(ctx, ts[k])
		if e != nil {
			t.Fatal(e)
		}
		vs[k] = v
	}
	for k, v := range values {
		vs[k] = v
	}
	return types.ObjectValueMust(ts, vs)
}
func stateFixture(t *testing.T, s resource.SchemaResponse, model types.Object) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: s.Schema}
	if d := st.Set(context.Background(), model); d.HasError() {
		t.Fatal(d)
	}
	return st
}
func nativeFixture(t *testing.T, value any) types.Dynamic {
	t.Helper()
	v, e := readNative(context.Background(), value, nil)
	if e != nil {
		t.Fatal(e)
	}
	return types.DynamicValue(v)
}
func noErrors(t *testing.T, d diag.Diagnostics) {
	t.Helper()
	if d.HasError() {
		t.Fatal(d)
	}
}

func TestConfigLifecycleUsesNativeStateAndStableIdentity(t *testing.T) {
	ctx := context.Background()
	var saved document
	deleted := false
	version := "revision-1"
	var writes []document
	r, s := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "POST", "PUT":
			var b document
			if e := json.NewDecoder(req.Body).Decode(&b); e != nil {
				t.Error(e)
			}
			writes = append(writes, b)
			saved = b
			if req.Method == "PUT" {
				version = "revision-2"
			}
			_, _ = fmt.Fprintf(w, `{"id":"config-id","version_id":%q}`, version)
		case "GET":
			if deleted {
				w.WriteHeader(404)
				_, _ = w.Write([]byte(`{"message":"absent"}`))
				return
			}
			doc, _ := json.Marshal(saved["config"])
			_, _ = fmt.Fprintf(w, `{"id":"config-id","workspace_id":"workspace","name":"test","config":%q,"version_id":%q}`, string(doc), version)
		case "DELETE":
			deleted = true
			_, _ = w.Write([]byte(`{"success":true}`))
		}
	})
	doc := map[string]any{"provider": "openai", "retry": map[string]any{"attempts": json.Number("1")}, "empty": []any{}, "disabled": false, "nullable": nil}
	model := modelFixture(t, r, map[string]attr.Value{"id": types.StringUnknown(), "name": types.StringValue("test"), "workspace_id": types.StringValue("workspace"), "config": nativeFixture(t, doc), "version_id": types.StringUnknown()})
	st := stateFixture(t, s, model)
	var create resource.CreateResponse
	create.State = tfsdk.State{Schema: s.Schema}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &create)
	noErrors(t, create.Diagnostics)
	var current types.Object
	noErrors(t, create.State.Get(ctx, &current))
	if !current.Attributes()["config"].Equal(model.Attributes()["config"]) {
		t.Fatal("native config types changed on read")
	}
	id, _ := resourceIDs(current)
	if id != "config-id" || current.Attributes()["version_id"].(types.String).ValueString() != "revision-1" {
		t.Fatal("receipt identity/version not persisted")
	}
	doc["retry"] = map[string]any{"attempts": json.Number("0")}
	values := current.Attributes()
	values["config"] = nativeFixture(t, doc)
	values["version_id"] = types.StringUnknown()
	plan := types.ObjectValueMust(current.AttributeTypes(ctx), values)
	st = stateFixture(t, s, plan)
	var update resource.UpdateResponse
	update.State = tfsdk.State{Schema: s.Schema}
	r.Update(ctx, resource.UpdateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &update)
	noErrors(t, update.Diagnostics)
	noErrors(t, update.State.Get(ctx, &current))
	id, _ = resourceIDs(current)
	if id != "config-id" || current.Attributes()["version_id"].(types.String).ValueString() != "revision-2" {
		t.Fatal("revision replaced resource identity")
	}
	if len(writes) != 2 || len(writes[1]["config"].(map[string]any)) != 5 {
		t.Fatal("update did not send the complete desired document")
	}
	var read resource.ReadResponse
	read.State = tfsdk.State{Schema: s.Schema}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
	noErrors(t, read.Diagnostics)
	if !read.State.Raw.Equal(update.State.Raw) {
		t.Fatal("subsequent refresh is not stable")
	}
	var destroy resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &destroy)
	noErrors(t, destroy.Diagnostics)
	if !deleted {
		t.Fatal("fixture not deleted")
	}
}

func TestConfigValidationRejectsJSONAndNestedCredentials(t *testing.T) {
	r := &gatewayResource{definition: definitions()[0]}
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	for _, value := range []attr.Value{types.DynamicValue(types.StringValue(`{"provider":"openai"}`)), nativeFixture(t, map[string]any{"targets": []any{map[string]any{"headers": map[string]any{"Authorization": "private"}}}}), types.DynamicValue(types.StringNull())} {
		model := modelFixture(t, r, map[string]attr.Value{"name": types.StringValue("test"), "workspace_id": types.StringValue("workspace"), "config": value})
		st := stateFixture(t, s, model)
		var response resource.ValidateConfigResponse
		r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s.Schema, Raw: st.Raw}}, &response)
		if !response.Diagnostics.HasError() {
			t.Fatal("unsafe/non-native config accepted")
		}
		if strings.Contains(fmt.Sprint(response.Diagnostics), "private") {
			t.Fatal("credential leaked in diagnostic")
		}
	}
}

func TestConfigRefreshDoesNotTreatAuthorizationFailureAsAbsence(t *testing.T) {
	r, s := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"message":"credential-should-not-leak"}`))
	})
	model := modelFixture(t, r, map[string]attr.Value{"id": types.StringValue("id"), "name": types.StringValue("test"), "workspace_id": types.StringValue("workspace"), "config": nativeFixture(t, map[string]any{})})
	st := stateFixture(t, s, model)
	resp := resource.ReadResponse{State: st}
	r.Read(context.Background(), resource.ReadRequest{State: st}, &resp)
	if !resp.Diagnostics.HasError() || resp.State.Raw.IsNull() {
		t.Fatal("authorization failure lost resource state")
	}
	if strings.Contains(fmt.Sprint(resp.Diagnostics), "credential-should-not-leak") {
		t.Fatal("SDK response leaked through diagnostics")
	}
}

func TestUnknownNestedConfigNeverReachesSDK(t *testing.T) {
	writes := 0
	r, s := configFixture(t, func(w http.ResponseWriter, _ *http.Request) { writes++; _, _ = w.Write([]byte(`{}`)) })
	payload := types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"provider": types.StringType}, map[string]attr.Value{"provider": types.StringUnknown()}))
	model := modelFixture(t, r, map[string]attr.Value{"name": types.StringValue("test"), "workspace_id": types.StringValue("workspace"), "config": payload})
	st := stateFixture(t, s, model)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &resp)
	if writes != 0 || !resp.Diagnostics.HasError() {
		t.Fatal("unknown input escaped to API")
	}
}

func TestRefreshKeepsNewRemoteRoutingKeysAsDrift(t *testing.T) {
	r, s := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"id","name":"test","workspace_id":"workspace","config":{"provider":"openai","new_field":false}}`))
	})
	model := modelFixture(t, r, map[string]attr.Value{"id": types.StringValue("id"), "name": types.StringValue("test"), "workspace_id": types.StringValue("workspace"), "config": nativeFixture(t, map[string]any{"provider": "openai"})})
	st := stateFixture(t, s, model)
	resp := resource.ReadResponse{State: st}
	r.Read(context.Background(), resource.ReadRequest{State: st}, &resp)
	noErrors(t, resp.Diagnostics)
	var current types.Object
	noErrors(t, resp.State.Get(context.Background(), &current))
	if current.Attributes()["config"].Equal(model.Attributes()["config"]) {
		t.Fatal("external config addition was hidden")
	}
}

func TestImportRejectsPlaintextRemoteRouting(t *testing.T) {
	r, s := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"id","name":"test","config":{"api_key":"do-not-store"}}`))
	})
	resp := resource.ImportStateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "id"}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("plaintext routing imported")
	}
	if strings.Contains(fmt.Sprint(resp.Diagnostics), "do-not-store") {
		t.Fatal("import leaked remote secret")
	}
}

func TestResolvedCredentialIsBlockedAtApplyBeforeSDK(t *testing.T) {
	writes := 0
	r, s := configFixture(t, func(w http.ResponseWriter, _ *http.Request) { writes++; _, _ = w.Write([]byte(`{}`)) })
	model := modelFixture(t, r, map[string]attr.Value{"name": types.StringValue("test"), "workspace_id": types.StringValue("workspace"), "config": nativeFixture(t, map[string]any{"client_secret": "resolved-secret"})})
	st := stateFixture(t, s, model)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	// Apply must enforce this even if the value was unknown during validation.
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &resp)
	if writes != 0 || !resp.Diagnostics.HasError() {
		t.Fatal("resolved credentials escaped the apply guard")
	}
}

func TestCreateReadFailurePersistsIdentityAndOneTimeSecret(t *testing.T) {
	r, s := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			_, _ = w.Write([]byte(`{"id":"owned-key","key":"one-time-material"}`))
			return
		}
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"message":"not-authorized"}`))
	})
	for _, d := range definitions() {
		if d.name == "service_api_key" {
			r.definition = d
		}
	}
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	scopes, ds := types.SetValueFrom(context.Background(), types.StringType, []string{"completions.write"})
	noErrors(t, ds)
	model := modelFixture(t, r, map[string]attr.Value{"name": types.StringValue("test"), "workspace_id": types.StringValue("workspace"), "scopes": scopes, "key": types.StringUnknown()})
	st := stateFixture(t, s, model)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("refresh failure was not reported")
	}
	var saved types.Object
	noErrors(t, resp.State.Get(context.Background(), &saved))
	if saved.Attributes()["id"].(types.String).ValueString() != "owned-key" || saved.Attributes()["key"].(types.String).ValueString() != "one-time-material" {
		t.Fatal("refresh failure discarded the creation receipt")
	}
}

func TestMaskedAndOneTimeSecretsNeverComeFromGET(t *testing.T) {
	ctx := context.Background()
	for _, d := range definitions() {
		if d.name != "integration" && d.name != "service_api_key" && d.name != "deployment" {
			continue
		}
		r := &gatewayResource{definition: d}
		model := modelFixture(t, r, nil)
		values := model.Attributes()
		for k, f := range d.fields {
			if f.retained || (f.computed && f.sensitive) {
				if f.kind == "object" {
					values[k] = nativeFixture(t, map[string]any{"owned": "secret"})
				} else {
					values[k] = types.StringValue("owned-secret")
				}
			}
		}
		model = types.ObjectValueMust(model.AttributeTypes(ctx), values)
		var ds diag.Diagnostics
		read := r.mapState(ctx, model, document{"key": "****", "client_auth": "****", "credentials": map[string]any{"masked": true}, "configurations": map[string]any{"masked": true}}, nil, &ds)
		noErrors(t, ds)
		for k, f := range d.fields {
			if (f.retained || (f.computed && f.sensitive)) && !read.Attributes()[k].Equal(model.Attributes()[k]) {
				t.Fatalf("masked read overwrote %s.%s", d.name, k)
			}
		}
	}
}

func TestGuardrailParametersAreSeparateSensitiveDesiredInput(t *testing.T) {
	ctx := context.Background()
	var d definition
	for _, candidate := range definitions() {
		if candidate.name == "guardrail" {
			d = candidate
		}
	}
	r := &gatewayResource{definition: d}
	model := modelFixture(t, r, nil)
	checkType := model.Attributes()["checks"].(types.List).ElementType(ctx).(types.ObjectType)
	check := types.ObjectValueMust(checkType.AttrTypes, map[string]attr.Value{"id": types.StringValue("check-id"), "name": types.StringUnknown(), "is_enabled": types.BoolValue(false)})
	checks, ds := types.ListValue(checkType, []attr.Value{check})
	noErrors(t, ds)
	values := model.Attributes()
	values["name"] = types.StringValue("test")
	values["workspace_id"] = types.StringValue("workspace")
	values["checks"] = checks
	values["check_parameters"] = nativeFixture(t, map[string]any{"check-id": map[string]any{"credential": "desired-secret", "enabled": false}})
	model = types.ObjectValueMust(model.AttributeTypes(ctx), values)
	var diags diag.Diagnostics
	body := r.body(model, false, &diags)
	noErrors(t, diags)
	if _, exists := body["check_parameters"]; exists {
		t.Fatal("provider parameter attribute escaped to SDK root")
	}
	item := body["checks"].([]any)[0].(map[string]any)
	if item["is_enabled"] != false || item["parameters"].(map[string]any)["enabled"] != false {
		t.Fatal("explicit false lost")
	}
	read := r.mapState(ctx, model, document{"checks": []any{map[string]any{"id": "check-id", "is_enabled": false, "parameters": map[string]any{"credential": "masked"}}}}, nil, &diags)
	noErrors(t, diags)
	if !read.Attributes()["check_parameters"].Equal(model.Attributes()["check_parameters"]) {
		t.Fatal("GET masked desired parameters")
	}
	values = model.Attributes()
	values["check_parameters"] = nativeFixture(t, map[string]any{"unowned-id": map[string]any{"value": "secret"}})
	model = types.ObjectValueMust(model.AttributeTypes(ctx), values)
	diags = nil
	r.body(model, false, &diags)
	if !diags.HasError() {
		t.Fatal("unowned parameter object was silently discarded")
	}
}
