package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func resourceDefinition(t *testing.T, name string) definition {
	t.Helper()
	for _, d := range definitions() {
		if d.name == name {
			return d
		}
	}
	t.Fatalf("missing definition %s", name)
	return definition{}
}

func TestRoutingCredentialPatternsAndReferences(t *testing.T) {
	for _, key := range []string{"openai_api_key", "anthropic_api_key", "azure_api_key", "x-portkey-api-key", "azure_ad_token", "vertex_service_account_json", "aws_secret_key", "gcp_service_account", "clientSecret", "privateKey", "Authorization"} {
		t.Run(key, func(t *testing.T) {
			if !routingCredentials(map[string]any{"targets": []any{map[string]any{key: "never-log-this"}}}) {
				t.Fatal("embedded credential accepted")
			}
		})
	}
	for _, key := range []string{"virtual_key", "secret_reference_id", "provider", "key", "metadata_key", "max_tokens"} {
		if routingCredentials(map[string]any{key: "reference"}) {
			t.Errorf("nonsecret reference %s rejected", key)
		}
	}
}

func TestApplyHonorsPlanAndRefreshExposesNormalization(t *testing.T) {
	ctx := context.Background()
	r, s := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" || req.Method == "PUT" {
			_, _ = w.Write([]byte(`{"id":"config-id","version_id":"revision"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"config-id","workspace_id":"workspace","name":"test","config":{"retry":{"attempts":1},"server_default":true},"version_id":"revision"}`))
	})
	plan := modelFixture(t, r, map[string]attr.Value{"id": types.StringUnknown(), "workspace_id": types.StringValue("workspace"), "name": types.StringValue("test"), "config": nativeFixture(t, map[string]any{"retry": map[string]any{"attempts": json.Number("1")}})})
	st := stateFixture(t, s, plan)
	create := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &create)
	noErrors(t, create.Diagnostics)
	var applied types.Object
	noErrors(t, create.State.Get(ctx, &applied))
	if !applied.Attributes()["config"].Equal(plan.Attributes()["config"]) {
		t.Fatal("apply changed a planned document")
	}
	update := resource.UpdateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: create.State.Raw}}, &update)
	noErrors(t, update.Diagnostics)
	noErrors(t, update.State.Get(ctx, &applied))
	if !applied.Attributes()["config"].Equal(plan.Attributes()["config"]) {
		t.Fatal("update changed a planned document")
	}
	read := resource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
	noErrors(t, read.Diagnostics)
	var refreshed types.Object
	noErrors(t, read.State.Get(ctx, &refreshed))
	if refreshed.Attributes()["config"].Equal(plan.Attributes()["config"]) {
		t.Fatal("refresh hid remote additions")
	}
}

func TestBindingOwnsExplicitPairWithoutOverridingGlobalOrOtherAccess(t *testing.T) {
	for _, mcp := range []bool{false, true} {
		t.Run(fmt.Sprint(mcp), func(t *testing.T) {
			enabled, missing := false, false
			writes := 0
			r, _ := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
				if !strings.Contains(req.URL.Path, "/workspaces") {
					if missing {
						w.WriteHeader(404)
						_, _ = w.Write([]byte(`{"message":"absent"}`))
						return
					}
					_, _ = w.Write([]byte(`{"id":"parent","global_workspace_access":{"enabled":true},"global_workspace_access_settings":{"enabled":true}}`))
					return
				}
				if req.Method == "GET" {
					key := "data"
					if mcp {
						key = "workspaces"
					}
					_, _ = fmt.Fprintf(w, `{%q:[{"id":"owned","enabled":%t},{"id":"other","enabled":true}],"global_workspace_access":{"enabled":true}}`, key, enabled)
					return
				}
				writes++
				var b document
				if err := json.NewDecoder(req.Body).Decode(&b); err != nil {
					t.Error(err)
				}
				if b["override_existing_workspace_access"] != false {
					t.Error("merge must be explicit")
				}
				if _, ok := b["global_workspace_access"]; ok {
					t.Error("global access mutated")
				}
				items := b["workspaces"].([]any)
				if len(items) != 1 || items[0].(map[string]any)["id"] != "owned" {
					t.Error("unowned workspace mutated")
				}
				item := items[0].(map[string]any)
				if mcp {
					if _, ok := b["create_default_provider"]; ok {
						t.Error("unmodeled MCP field")
					}
					if _, ok := item["create_default_provider"]; ok {
						t.Error("unmodeled MCP item field")
					}
				}
				enabled = item["enabled"].(bool)
				_, _ = w.Write([]byte(`{"success":true}`))
			})
			d := bindingDefinition(mcp)
			ctx := context.Background()
			body := document{"integration_id": "parent", "workspace_id": "owned"}
			_, err := d.create(ctx, r.client, body)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.read(ctx, r.client, "parent/owned", "")
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.create(ctx, r.client, body)
			if err == nil || !strings.Contains(gatewayError(err), "409") {
				t.Fatal("existing binding was adopted", err)
			}
			if writes != 1 {
				t.Fatal("existing binding triggered a write")
			}
			if err = d.delete(ctx, r.client, "parent/owned", ""); err != nil {
				t.Fatal(err)
			}
			if enabled || writes != 2 {
				t.Fatal("destroy did not disable just the pair")
			}
			_, err = d.read(ctx, r.client, "parent/owned", "")
			if !aisec.IsNotFound(err) {
				t.Fatal("disabled explicit pair treated as enabled through global access")
			}
			missing = true
			_, err = d.read(ctx, r.client, "parent/owned", "")
			if !aisec.IsNotFound(err) {
				t.Fatal("absent parent was not absent")
			}
		})
	}
}

func TestDeploymentBooleanAndArchivedRefresh(t *testing.T) {
	for _, flag := range []string{"0", "1"} {
		t.Run(flag, func(t *testing.T) {
			r, s := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprintf(w, `{"id":"deployment","name":"test","type":"non_production","is_default":%s,"status":"archived"}`, flag)
			})
			r.definition = resourceDefinition(t, "deployment")
			r.Schema(context.Background(), resource.SchemaRequest{}, &s)
			m := modelFixture(t, r, map[string]attr.Value{"id": types.StringValue("deployment"), "is_default": types.BoolUnknown()})
			var diags diag.Diagnostics
			mapped := r.mapState(context.Background(), m, document{"is_default": json.Number(flag)}, nil, &diags)
			noErrors(t, diags)
			if mapped.Attributes()["is_default"].(types.Bool).ValueBool() != (flag == "1") {
				t.Fatal("0/1 boolean decoding")
			}
			read := resource.ReadResponse{State: stateFixture(t, s, m)}
			r.Read(context.Background(), resource.ReadRequest{State: read.State}, &read)
			noErrors(t, read.Diagnostics)
			if !read.State.Raw.IsNull() {
				t.Fatal("archived deployment retained in state")
			}
		})
	}
}

func TestScopedImportRejectsWrongWorkspaceAndMalformedIdentifier(t *testing.T) {
	r, s := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/server") {
			_, _ = w.Write([]byte(`{"id":"server","name":"test","mcp_integration_id":"parent"}`))
			return
		}
		if req.URL.Query().Get("workspace_id") != "correct" {
			_, _ = w.Write([]byte(`{"data":[],"total":0}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"server"}],"total":1}`))
	})
	r.definition = resourceDefinition(t, "mcp_server")
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	for _, id := range []string{"server", "wrong/server"} {
		resp := resource.ImportStateResponse{State: tfsdk.State{Schema: s.Schema}}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Fatalf("invalid import %s accepted", id)
		}
	}
	resp := resource.ImportStateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "correct/server"}, &resp)
	noErrors(t, resp.Diagnostics)
	if err := r.validateImportScope(context.Background(), document{"workspace_id": "wrong"}, "server", "correct"); err == nil {
		t.Fatal("detail mismatch accepted")
	}
}

func TestDiscoveryPaginationAndSafeTypedMetadata(t *testing.T) {
	r, _ := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("current_page") != "2" || req.URL.Query().Get("page_size") != "1" || req.URL.Query().Get("workspace_id") != "workspace" {
			t.Error("paging or scope not forwarded", req.URL.Query())
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"key-id","name":"test","version_id":"revision","is_default":0,"enabled":true,"key":"do-not-store","config":{"api_key":"do-not-store"}}],"total":12}`))
	})
	d := gatewayDataSource{definition: resourceDefinition(t, "service_api_key"), client: r.client}
	var s datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &s)
	ts := s.Schema.Type().(types.ObjectType).AttrTypes
	values := map[string]attr.Value{}
	for k, typ := range ts {
		v, err := nullNative(context.Background(), typ)
		if err != nil {
			t.Fatal(err)
		}
		values[k] = v
	}
	values["workspace_id"] = types.StringValue("workspace")
	values["current_page"] = types.Int64Value(2)
	values["page_size"] = types.Int64Value(1)
	model := types.ObjectValueMust(ts, values)
	st := tfsdk.State{Schema: s.Schema}
	noErrors(t, st.Set(context.Background(), model))
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
	d.Read(context.Background(), datasource.ReadRequest{Config: tfsdk.Config{Schema: s.Schema, Raw: st.Raw}}, &resp)
	noErrors(t, resp.Diagnostics)
	var actual types.Object
	noErrors(t, resp.State.Get(context.Background(), &actual))
	if actual.Attributes()["total_count"].(types.Int64).ValueInt64() != 12 {
		t.Fatal("total lost")
	}
	item := actual.Attributes()["items"].(types.List).Elements()[0].(types.Object).Attributes()
	if item["version_id"].(types.String).ValueString() != "revision" || item["is_default"].(types.Bool).ValueBool() || !item["enabled"].(types.Bool).ValueBool() {
		t.Fatal("typed summary lost")
	}
	encoded, err := nativeJSON(actual)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(encoded)
	if strings.Contains(string(b), "do-not-store") {
		t.Fatal("discovery leaked credentials or document")
	}
}

func TestTypedChecksFillUnknownChildrenAndKeepKnownLeaves(t *testing.T) {
	ctx := context.Background()
	r := &gatewayResource{definition: resourceDefinition(t, "guardrail")}
	typ := r.attributes()["checks"].GetType().(types.ListType).ElemType.(types.ObjectType)
	planned := types.ObjectValueMust(typ.AttrTypes, map[string]attr.Value{"id": types.StringValue("check"), "name": types.StringUnknown(), "is_enabled": types.BoolValue(false)})
	prior := types.ListValueMust(typ, []attr.Value{planned})
	remote, err := readChecks(ctx, []any{map[string]any{"id": "check", "name": "server name", "is_enabled": true, "parameters": map[string]any{"api_key": "masked"}}}, prior)
	if err != nil {
		t.Fatal(err)
	}
	applied := honorPlanned(prior, remote).(types.List).Elements()[0].(types.Object).Attributes()
	if applied["name"].(types.String).ValueString() != "server name" || applied["is_enabled"].(types.Bool).ValueBool() {
		t.Fatal("unknown child not filled or known child changed")
	}
	if _, ok := applied["parameters"]; ok {
		t.Fatal("masked check parameters imported")
	}
	actionType := r.attributes()["actions"].GetType().(types.ObjectType)
	template := types.ObjectNull(actionType.AttrTypes)
	feedback := map[string]any{"feedback": map[string]any{"value": json.Number("0"), "weight": json.Number("1"), "metadata": ""}}
	mapped, err := readTypedObject(ctx, map[string]any{"deny": false, "async": false, "on_success": feedback, "on_fail": feedback, "undeclared": "ignored"}, template)
	if err != nil {
		t.Fatal(err)
	}
	if mapped.(types.Object).Attributes()["deny"].(types.Bool).ValueBool() {
		t.Fatal("typed false lost")
	}
}

func TestSDKShapeDiagnosticNamesFieldWithoutInputValues(t *testing.T) {
	type request struct {
		Defaults struct {
			AllowConfigOverride bool `json:"allow_config_override"`
		} `json:"defaults"`
	}
	_, err := writeSDK(context.Background(), document{"defaults": map[string]any{"allow_config_override": "sensitive-invalid-value"}}, func(context.Context, request) (*document, error) {
		t.Fatal("invalid input reached network")
		return nil, nil
	})
	message := gatewayError(err)
	if err == nil || !strings.Contains(message, "defaults.allow_config_override") || strings.Contains(message, "sensitive-invalid-value") {
		t.Fatal("shape diagnostic missing field or exposed value", message)
	}
}
