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
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
		if req.URL.Query().Get("current_page") != "1" || req.URL.Query().Get("page_size") != "1" || req.URL.Query().Get("workspace_id") != "workspace" {
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
	honored, err := honorPlanned(prior, remote)
	if err != nil {
		t.Fatal(err)
	}
	applied := honored.(types.List).Elements()[0].(types.Object).Attributes()
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

func TestProviderUsageSettingsSeparatePolicyMetadataFromManagedLeaves(t *testing.T) {
	ctx := context.Background()
	r := &gatewayResource{definition: resourceDefinition(t, "provider")}
	typ := r.attributes()["usage_limits"].GetType().(types.ObjectType)
	values := map[string]attr.Value{}
	for k, kind := range typ.AttrTypes {
		value, err := kind.ValueFromTerraform(ctx, tftypes.NewValue(kind.TerraformType(ctx), tftypes.UnknownValue))
		if err != nil {
			t.Fatal(err)
		}
		values[k] = value
	}
	values["type"] = types.StringValue("tokens")
	values["credit_limit"] = types.Int64Value(100000)
	values["alert_threshold"] = types.Int64Value(0)
	planned := types.ObjectValueMust(typ.AttrTypes, values)
	body, err := writeSettings(planned)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != 3 || body["alert_threshold"] != int64(0) {
		t.Fatal("unconfigured defaults sent or explicit zero lost")
	}
	remote, err := readTypedObject(ctx, map[string]any{"id": "server-only-policy-id", "type": "tokens", "credit_limit": json.Number("100000"), "alert_threshold": json.Number("0"), "periodic_reset": "monthly"}, planned)
	if err != nil {
		t.Fatal(err)
	}
	honored, err := honorPlanned(planned, remote)
	if err != nil {
		t.Fatal(err)
	}
	mapped := honored.(types.Object).Attributes()
	if mapped["periodic_reset"].(types.String).ValueString() != "monthly" || mapped["alert_threshold"].(types.Int64).ValueInt64() != 0 {
		t.Fatal("default or explicit zero not mapped")
	}
	if _, ok := mapped["id"]; ok {
		t.Fatal("server policy metadata became managed config")
	}
}

func TestPartialDynamicPlanKeepsShapeAndKnownLeaves(t *testing.T) {
	ctx := context.Background()
	targetType := types.ObjectType{AttrTypes: map[string]attr.Type{"provider": types.StringType, "weight": types.Int64Type}}
	target := types.ObjectValueMust(targetType.AttrTypes, map[string]attr.Value{"provider": types.StringUnknown(), "weight": types.Int64Value(1)})
	mapValue := types.MapValueMust(types.StringType, map[string]attr.Value{"computed": types.StringUnknown(), "known": types.StringValue("keep")})
	setType := types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType, "computed": types.StringType}}
	set := types.SetValueMust(setType, []attr.Value{types.ObjectValueMust(setType.AttrTypes, map[string]attr.Value{"id": types.StringValue("a"), "computed": types.StringUnknown()})})
	tuple := types.TupleValueMust([]attr.Type{targetType}, []attr.Value{target})
	object := types.ObjectValueMust(map[string]attr.Type{"targets": tuple.Type(ctx), "mapping": mapValue.Type(ctx), "members": set.Type(ctx)}, map[string]attr.Value{"targets": tuple, "mapping": mapValue, "members": set})
	plan := types.DynamicValue(object)
	remote, err := readNative(ctx, map[string]any{"targets": []any{map[string]any{"provider": "@slug", "weight": json.Number("2"), "server_added": true}}, "mapping": map[string]any{"computed": "filled", "known": "normalized", "server_added": "extra"}, "members": []any{map[string]any{"id": "a", "computed": "filled", "server_added": true}}, "server_added": true}, plan)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := honorPlanned(plan, remote)
	if err != nil {
		t.Fatal(err)
	}
	attrs := applied.(types.Dynamic).UnderlyingValue().(types.Object).Attributes()
	if len(attrs) != 3 {
		t.Fatal("remote-only root key became planned")
	}
	got := attrs["targets"].(types.Tuple).Elements()[0].(types.Object).Attributes()
	if got["provider"].(types.String).ValueString() != "@slug" || got["weight"].(types.Int64).ValueInt64() != 1 || len(got) != 2 {
		t.Fatal("partial tuple/object plan changed known leaves or shape")
	}
	mapping := attrs["mapping"].(types.Map).Elements()
	if mapping["known"].(types.String).ValueString() != "keep" || mapping["computed"].(types.String).ValueString() != "filled" || len(mapping) != 2 {
		t.Fatal("partial map plan changed")
	}
	if !remote.(types.Dynamic).UnderlyingValue().(types.Object).Attributes()["server_added"].(types.Bool).ValueBool() {
		t.Fatal("read did not expose remote additions")
	}
}

func TestCredentialValuesHeadersAndNonsecretTokens(t *testing.T) {
	for _, value := range []any{map[string]any{"headers": []any{map[string]any{"name": "Authorization", "value": "test-placeholder"}}}, map[string]any{"neutral": "sk-abcdefghijklmnopqrstuvwxyz12345"}, map[string]any{"neutral": "AKIAABCDEFGHIJKLMNOP"}, map[string]any{"neutral": "-----BEGIN PRIVATE KEY-----\nplaceholder\n-----END PRIVATE KEY-----"}} {
		if !routingCredentials(value) {
			t.Fatal("recognizable credential or header accepted")
		}
	}
	if routingCredentials(map[string]any{"override_params": map[string]any{"pad_token": "<pad>", "eos_token": "<eos>"}}) {
		t.Fatal("nonsecret model tokens rejected")
	}
}

func TestNullGuardrailActionsAndTimestampEquality(t *testing.T) {
	r := &gatewayResource{definition: resourceDefinition(t, "guardrail")}
	prior := modelFixture(t, r, map[string]attr.Value{})
	var d diag.Diagnostics
	actual := r.mapState(context.Background(), prior, document{"actions": nil}, nil, &d)
	noErrors(t, d)
	if !actual.Attributes()["actions"].IsNull() {
		t.Fatal("null actions not retained as readable null")
	}
	old := types.StringValue("2027-01-01T01:00:00+01:00")
	if equivalentTimestamp("2027-01-01T00:00:00.000Z", old) != old.ValueString() {
		t.Fatal("equivalent timestamp created drift")
	}
	if equivalentTimestamp("2027-01-02T00:00:00Z", old) == old.ValueString() {
		t.Fatal("real expiry drift hidden")
	}
}

func TestNonemptyIntegrationSecretMappingsUseTypedSDKContracts(t *testing.T) {
	for _, kind := range []string{"integration", "mcp_integration"} {
		t.Run(kind, func(t *testing.T) {
			r, _ := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
				var body document
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				mappings := body["secret_mappings"].([]any)
				mapping := mappings[0].(map[string]any)
				if mapping["target_field"] != "api_key" || mapping["secret_reference_id"] != "reference-uuid" || mapping["value_format"] != "string" {
					t.Error("secret reference mapping changed at SDK seam")
				}
				if _, ok := mapping["value"]; ok {
					t.Error("inline secret added to reference")
				}
				_, _ = w.Write([]byte(`{"id":"integration","slug":"integration"}`))
			})
			def := resourceDefinition(t, kind)
			body := document{"name": "test", "secret_mappings": []any{map[string]any{"target_field": "api_key", "secret_reference_id": "reference-uuid", "secret_key": "credential", "value_format": "string"}}}
			if kind == "integration" {
				body["ai_provider_id"] = "family"
			} else {
				body["url"], body["auth_type"], body["transport"] = "https://example.test/mcp", "none", "http"
			}
			_, err := def.create(context.Background(), r.client, body)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDefaultDeploymentRequiresExplicitTrueAndDoesNotRotate(t *testing.T) {
	for _, desired := range []bool{false, true} {
		t.Run(fmt.Sprint(desired), func(t *testing.T) {
			r, s := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
				if req.Method == "POST" {
					var body document
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body["is_default"] != desired || body["organisation_id"] != "123" {
						t.Error("explicit default or declared organisation not sent")
					}
					if _, ok := body["rotate_auth"]; ok {
						t.Error("auth rotation sent automatically")
					}
					_, _ = w.Write([]byte(`{"id":"deployment"}`))
					return
				}
				flag := 0
				if desired {
					flag = 1
				}
				_, _ = fmt.Fprintf(w, `{"id":"deployment","name":"test","type":"non_production","is_default":%d}`, flag)
			})
			r.definition = resourceDefinition(t, "deployment")
			r.Schema(context.Background(), resource.SchemaRequest{}, &s)
			planned := modelFixture(t, r, map[string]attr.Value{"name": types.StringValue("test"), "type": types.StringValue("non_production"), "is_default": types.BoolValue(desired)})
			st := stateFixture(t, s, planned)
			resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &resp)
			noErrors(t, resp.Diagnostics)
		})
	}
}

func TestSecretWorkspaceDriftLimitationProducesActionableWarning(t *testing.T) {
	r, s := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"reference","name":"test","manager_type":"aws_sm","secret_path":"path","allow_all_workspaces":false}`))
	})
	r.definition = resourceDefinition(t, "secret_reference")
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	m := modelFixture(t, r, map[string]attr.Value{"id": types.StringValue("reference"), "allowed_workspaces": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("workspace")})})
	response := resource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Read(context.Background(), resource.ReadRequest{State: stateFixture(t, s, m)}, &response)
	noErrors(t, response.Diagnostics)
	if response.Diagnostics.WarningsCount() != 1 {
		t.Fatal("missing access policy was silently treated as verified")
	}
	var actual types.Object
	noErrors(t, response.State.Get(context.Background(), &actual))
	if !actual.Attributes()["allowed_workspaces"].Equal(m.Attributes()["allowed_workspaces"]) {
		t.Fatal("unavailable workspace policy erased")
	}
}

func TestUnknownSetMembersResolveWithoutOrderDependentAmbiguity(t *testing.T) {
	remote := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	for _, elements := range [][]attr.Value{
		{types.StringUnknown(), types.StringValue("a")},
		{types.StringValue("a"), types.StringUnknown()},
		{types.StringUnknown(), types.StringUnknown()},
	} {
		got, err := honorPlanned(types.SetValueMust(types.StringType, elements), remote)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(remote) {
			t.Fatal("unknown set did not resolve to known remote members")
		}
	}
}

func TestApplyUsesReceiptForOmittedAccessListAndRejectsUnresolvedState(t *testing.T) {
	r := &gatewayResource{definition: resourceDefinition(t, "secret_reference")}
	plan := modelFixture(t, r, map[string]attr.Value{"allowed_workspaces": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringUnknown()})})
	var d diag.Diagnostics
	got := r.mapAppliedState(context.Background(), plan, document{}, document{"allowed_workspaces": []any{"a", "b"}}, &d)
	noErrors(t, d)
	expected := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	if !got.Attributes()["allowed_workspaces"].Equal(expected) {
		t.Fatal("write receipt not used for omitted access list")
	}
	d = nil
	r.mapAppliedState(context.Background(), plan, document{}, nil, &d)
	if !d.HasError() {
		t.Fatal("unresolved nested value allowed in final state")
	}
}

func TestSecretImportWarnsWhenWorkspacePolicyIsOmitted(t *testing.T) {
	r, schemaResponse := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"reference","name":"test","manager_type":"aws_sm","secret_path":"path","allow_all_workspaces":false}`))
	})
	r.definition = resourceDefinition(t, "secret_reference")
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	response := resource.ImportStateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "reference"}, &response)
	noErrors(t, response.Diagnostics)
	found := false
	for _, d := range response.Diagnostics {
		if d.Summary() == "Workspace access unavailable on import" {
			found = true
		}
	}
	if !found {
		t.Fatal("import did not warn about unavailable workspace policy")
	}
}

func TestAdditionalCredentialPrefixes(t *testing.T) {
	for _, value := range []string{"ghp_abcdefghijklmnopqrstuvwxyz", "github_pat_abcdefghijklmnopqrstuvwxyz", "xoxb-" + strings.Repeat("a", 24), "AIzaabcdefghijklmnopqrstuvwxyz123456", "glpat-abcdefghijklmnopqrstuvwxyz"} {
		if !routingCredentials(map[string]any{"neutral": value}) {
			t.Fatal("recognized secret prefix accepted")
		}
	}
	if routingCredentials(map[string]any{"start_token": "<start>", "stop_token": "<stop>"}) {
		t.Fatal("nonsecret model token rejected")
	}
}

func TestPartialWorkspaceListingsNeverProveBindingAbsence(t *testing.T) {
	for _, mcp := range []bool{false, true} {
		// Integrations declare workspaces; MCP declares data and accepts legacy workspaces.
		for _, key := range []string{"workspaces", "data"} {
			for _, metadata := range []string{`"total":2`, `"has_more":true`} {
				writes := 0
				r, _ := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
					if req.Method != "GET" {
						writes++
						t.Error("partial inventory caused a write")
						return
					}
					if !strings.Contains(req.URL.Path, "/workspaces") {
						_, _ = w.Write([]byte(`{"id":"parent"}`))
						return
					}
					_, _ = fmt.Fprintf(w, `{%q:[{"id":"other","enabled":true}],%s}`, key, metadata)
				})
				def := bindingDefinition(mcp)
				_, err := def.read(context.Background(), r.client, "parent/owned", "")
				if err == nil || aisec.IsNotFound(err) {
					t.Fatal("partial workspace page treated as authoritative absence", err)
				}
				_, err = def.create(context.Background(), r.client, document{"integration_id": "parent", "workspace_id": "owned"})
				if err == nil || writes != 0 {
					t.Fatal("partial inventory permitted create")
				}
				err = def.delete(context.Background(), r.client, "parent/owned", "")
				if err == nil || writes != 0 {
					t.Fatal("partial inventory permitted destroy write")
				}
			}
		}
	}
}

func TestDisabledOwnedPairConfirmsAbsenceOnPartialPage(t *testing.T) {
	for _, mcp := range []bool{false, true} {
		r, _ := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
			if req.Method != "GET" {
				t.Error("already disabled binding caused a write")
				return
			}
			if !strings.Contains(req.URL.Path, "/workspaces") {
				_, _ = w.Write([]byte(`{"id":"parent"}`))
				return
			}
			key := "workspaces"
			if mcp {
				key = "data"
			}
			_, _ = fmt.Fprintf(w, `{%q:[{"id":"owned","enabled":false}],"total":2,"has_more":true}`, key)
		})
		def := bindingDefinition(mcp)
		_, err := def.read(context.Background(), r.client, "parent/owned", "")
		if !aisec.IsNotFound(err) {
			t.Fatal("disabled owned pair not confirmed", err)
		}
		if err = def.delete(context.Background(), r.client, "parent/owned", ""); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOmittedEmptyOptionalStringsRemainCanonical(t *testing.T) {
	for kind, field := range map[string]string{"secret_reference": "secret_key", "usage_limit": "periodic_reset"} {
		r := &gatewayResource{definition: resourceDefinition(t, kind)}
		old := modelFixture(t, r, map[string]attr.Value{field: types.StringValue("")})
		var d diag.Diagnostics
		got := r.mapState(context.Background(), old, document{}, nil, &d)
		noErrors(t, d)
		if !got.Attributes()[field].Equal(types.StringValue("")) {
			t.Fatal("empty optional string produced repeat drift")
		}
	}
}

func TestRefreshShowsRemoteOptionalClearsButRetainsKnownOmission(t *testing.T) {
	for _, kind := range []string{"integration", "mcp_integration", "provider", "secret_reference", "deployment"} {
		r := &gatewayResource{definition: resourceDefinition(t, kind)}
		field := "description"
		if kind == "provider" {
			field = "note"
		}
		if kind == "deployment" {
			field = "tags"
		}
		if kind == "secret_reference" {
			field = "secret_key"
		}
		value := attr.Value(types.StringValue("old"))
		if field == "tags" {
			value = types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"old": types.BoolType}, map[string]attr.Value{"old": types.BoolValue(true)}))
		}
		prior := modelFixture(t, r, map[string]attr.Value{field: value})
		var d diag.Diagnostics
		got := r.mapState(context.Background(), prior, document{}, nil, &d)
		noErrors(t, d)
		if !got.Attributes()[field].IsNull() {
			t.Fatalf("%s.%s remote clear hidden", kind, field)
		}
	}
	r := &gatewayResource{definition: resourceDefinition(t, "secret_reference")}
	desired := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("workspace")})
	prior := modelFixture(t, r, map[string]attr.Value{"allowed_workspaces": desired})
	var d diag.Diagnostics
	got := r.mapState(context.Background(), prior, document{}, nil, &d)
	noErrors(t, d)
	if !got.Attributes()["allowed_workspaces"].Equal(desired) {
		t.Fatal("known access-list omission lost desired policy")
	}
}

func TestUnpagedDiscoveryWarnsAboutIncompleteInventory(t *testing.T) {
	d := gatewayDataSource{definition: resourceDefinition(t, "config")}
	d.definition.list = func(context.Context, *client, string, int64, int64) ([]document, int64, error) {
		return []document{{"id": "owned"}}, 2, nil
	}
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
	st := tfsdk.State{Schema: s.Schema}
	noErrors(t, st.Set(context.Background(), types.ObjectValueMust(ts, values)))
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
	d.Read(context.Background(), datasource.ReadRequest{Config: tfsdk.Config{Schema: s.Schema, Raw: st.Raw}}, &resp)
	noErrors(t, resp.Diagnostics)
	found := false
	for _, diagnostic := range resp.Diagnostics {
		if diagnostic.Summary() == "Gateway listing is incomplete" {
			found = true
		}
	}
	if !found {
		t.Fatal("unpaged discovery hid incomplete inventory")
	}
}

func TestOmittedEmptyCollectionsPreserveShapeAndNonemptyClearsDrift(t *testing.T) {
	r := &gatewayResource{definition: resourceDefinition(t, "integration")}
	empty := types.DynamicValue(types.TupleValueMust([]attr.Type{}, []attr.Value{}))
	nonempty := types.DynamicValue(types.TupleValueMust([]attr.Type{types.StringType}, []attr.Value{types.StringValue("reference")}))
	for _, value := range []attr.Value{empty, nonempty} {
		prior := modelFixture(t, r, map[string]attr.Value{"secret_mappings": value})
		var d diag.Diagnostics
		got := r.mapState(context.Background(), prior, document{}, nil, &d)
		noErrors(t, d)
		if value.Equal(empty) {
			if !got.Attributes()["secret_mappings"].Equal(empty) {
				t.Fatal("omitted empty collection changed HCL shape")
			}
		} else if !got.Attributes()["secret_mappings"].IsNull() {
			t.Fatal("nonempty collection clear was hidden")
		}
	}
}
