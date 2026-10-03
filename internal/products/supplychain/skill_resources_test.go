package supplychain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ag "github.com/cdot65/prisma-airs-go/aisec/agentguard"
	s "github.com/cdot65/prisma-airs-go/aisec/agentguard/schema"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const skillTestID = "12345678-1234-1234-1234-123456789abc"

func skillResourceFixture(t *testing.T, kind string, handler http.HandlerFunc) (*skillResource, resource.SchemaResponse) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	sdk, e := ag.NewClient(ag.Opts{ClientID: "client", ClientSecret: "secret", TsgID: "123", DataEndpoint: server.URL, MgmtEndpoint: server.URL, TokenEndpoint: server.URL + "/token", NumRetries: 1})
	if e != nil {
		t.Fatal(e)
	}
	r := &skillResource{kind: kind, clients: &Clients{Skills: sdk, TenantID: "123"}}
	var sr resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &sr)
	return r, sr
}
func skillModel(t *testing.T, r *skillResource, values map[string]attr.Value) types.Object {
	t.Helper()
	vs := map[string]attr.Value{}
	for k, a := range r.attributes() {
		vs[k] = types.StringNull()
		if a.GetType().Equal(types.BoolType) {
			vs[k] = types.BoolNull()
		} else if a.GetType().Equal(types.Int64Type) {
			vs[k] = types.Int64Null()
		} else if a.GetType().Equal(types.DynamicType) {
			vs[k] = types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}))
		}
	}
	for k, v := range values {
		vs[k] = v
	}
	return r.object(context.Background(), vs)
}
func skillState(t *testing.T, sr resource.SchemaResponse, v types.Object) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: sr.Schema}
	noSkillErrors(t, st.Set(context.Background(), v))
	return st
}
func noSkillErrors(t *testing.T, d diag.Diagnostics) {
	t.Helper()
	if d.HasError() {
		t.Fatal(d)
	}
}
func TestSkillScanningCreateCheckpointsIdentityBeforeFailedRefresh(t *testing.T) {
	r, sr := skillResourceFixture(t, "instance", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			_, _ = w.Write([]byte(`{"is_success":true,"tenant_id":"tenant","tsg_id":"123"}`))
			return
		}
		w.WriteHeader(404)
	})
	plan := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"tenant_id": types.StringValue("tenant"), "support_account_id": types.StringValue("support"), "created_by": types.StringValue("user"), "id": types.StringUnknown(), "tsg_id": types.StringUnknown()}))
	resp := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: plan.Raw}, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: plan.Raw}}, &resp)
	if !resp.Diagnostics.HasError() || resp.State.Raw.IsNull() {
		t.Fatal("successful create lost identity after failed refresh")
	}
	var v types.Object
	noSkillErrors(t, resp.State.Get(context.Background(), &v))
	if skillString(v.Attributes(), "id") != "tenant" {
		t.Fatal("checkpoint lost tenant")
	}
}
func TestSkillScanningRefreshRetainsStateOnForbiddenOrIncompleteInventory(t *testing.T) {
	for _, kind := range []string{"instance", "rule", "override"} {
		t.Run(kind, func(t *testing.T) {
			r, sr := skillResourceFixture(t, kind, func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(403)
				_, _ = w.Write([]byte(`{"message":"private-secret"}`))
			})
			values := map[string]attr.Value{"id": types.StringValue(skillTestID), "tsg_id": types.StringValue("123")}
			if kind == "rule" {
				values["rule_uuid"] = types.StringValue(skillTestID)
			}
			st := skillState(t, sr, skillModel(t, r, values))
			resp := resource.ReadResponse{State: st}
			r.Read(context.Background(), resource.ReadRequest{State: st}, &resp)
			if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(st.Raw) {
				t.Fatal("permission failure lost prior state")
			}
			if strings.Contains(fmt.Sprint(resp.Diagnostics), "private-secret") {
				t.Fatal("server body leaked")
			}
		})
	}
}
func TestSkillScanningInventoryContinuesPastClampedAndPageCountTotals(t *testing.T) {
	calls := 0
	r, _ := skillResourceFixture(t, "rule", func(w http.ResponseWriter, req *http.Request) {
		calls++
		if req.URL.Query().Get("skip") == "0" {
			_, _ = fmt.Fprintf(w, `{"rules":[{"uuid":%q}],"pagination":{"total_items":1}}`, skillTestID)
		} else {
			_, _ = w.Write([]byte(`{"rules":[],"pagination":{"total_items":0}}`))
		}
	})
	rows, e := allRules(context.Background(), r.clients.Skills)
	if e != nil || len(rows) != 1 || calls != 2 {
		t.Fatal("did not confirm full inventory with terminal empty page", e)
	}
}
func TestSkillScanningRepeatedInventoryNeverReportsAbsence(t *testing.T) {
	r, sr := skillResourceFixture(t, "override", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"skill_overrides":[{"uuid":%q}],"pagination":{"total_items":1}}`, skillTestID)
	})
	st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"id": types.StringValue("87654321-4321-4321-4321-cba987654321"), "tsg_id": types.StringValue("123")}))
	resp := resource.ReadResponse{State: st}
	r.Read(context.Background(), resource.ReadRequest{State: st}, &resp)
	if !resp.Diagnostics.HasError() || !resp.State.Raw.Equal(st.Raw) {
		t.Fatal("duplicate pages caused false absence")
	}
}
func TestSkillScanningDestroyChecksReceiptAndConfirmsAbsence(t *testing.T) {
	for _, scenario := range []string{"false receipt", "still present", "gone"} {
		t.Run(scenario, func(t *testing.T) {
			r, sr := skillResourceFixture(t, "instance", func(w http.ResponseWriter, req *http.Request) {
				if req.Method == "DELETE" {
					_, _ = fmt.Fprintf(w, `{"is_success":%t,"tenant_id":"tenant","tsg_id":"123"}`, scenario != "false receipt")
					return
				}
				if scenario == "gone" {
					w.WriteHeader(404)
					return
				}
				_, _ = w.Write([]byte(`{"tenant_id":"tenant","tsg_id":"123","support_account_id":"support"}`))
			})
			st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"id": types.StringValue("tenant"), "tsg_id": types.StringValue("123")}))
			var resp resource.DeleteResponse
			r.Delete(context.Background(), resource.DeleteRequest{State: st}, &resp)
			if resp.Diagnostics.HasError() != (scenario != "gone") {
				t.Fatal("deletion result is incorrect", resp.Diagnostics)
			}
		})
	}
}
func TestSkillScanningCrossTenantDestroyAndUpdateMakeNoRequests(t *testing.T) {
	for _, kind := range []string{"instance", "rule", "override"} {
		t.Run(kind, func(t *testing.T) {
			calls := 0
			r, sr := skillResourceFixture(t, kind, func(http.ResponseWriter, *http.Request) { calls++ })
			st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"id": types.StringValue(skillTestID), "tsg_id": types.StringValue("different")}))
			var resp resource.DeleteResponse
			r.Delete(context.Background(), resource.DeleteRequest{State: st}, &resp)
			if !resp.Diagnostics.HasError() || calls != 0 {
				t.Fatal("cross-tenant destroy escaped")
			}
			update := resource.UpdateResponse{State: st}
			r.Update(context.Background(), resource.UpdateRequest{State: st, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &update)
			if !update.Diagnostics.HasError() || calls != 0 {
				t.Fatal("cross-tenant update escaped")
			}
		})
	}
}
func TestSkillScanningUnknownInputsFailBeforeAuthentication(t *testing.T) {
	calls := 0
	r, sr := skillResourceFixture(t, "instance", func(http.ResponseWriter, *http.Request) { calls++ })
	st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"tenant_id": types.StringUnknown()}))
	resp := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: st.Raw}, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &resp)
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatal("unknown input escaped to API")
	}
}
func TestSkillScanningNativeNullFalseAndLargeNumbers(t *testing.T) {
	v, e := skillNative(context.Background(), map[string]any{"null": nil, "false": false, "number": json.Number("9007199254740993"), "empty": []any{}})
	if e != nil {
		t.Fatal(e)
	}
	m := v.(types.Object).Attributes()
	if !m["null"].IsNull() || m["false"].(types.Bool).ValueBool() || m["number"].(types.Number).ValueBigFloat().Text('f', 0) != "9007199254740993" || len(m["empty"].(types.Tuple).Elements()) != 0 {
		t.Fatal("native mapping lost value semantics")
	}
}

func TestSkillScanningDestroyNeverEchoesUpstreamErrors(t *testing.T) {
	for _, kind := range []string{"instance", "override"} {
		t.Run(kind, func(t *testing.T) {
			r, sr := skillResourceFixture(t, kind, func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(403)
				_, _ = w.Write([]byte(`{"message":"private-secret"}`))
			})
			st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"id": types.StringValue(skillTestID), "tsg_id": types.StringValue("123")}))
			var resp resource.DeleteResponse
			r.Delete(context.Background(), resource.DeleteRequest{State: st}, &resp)
			if !resp.Diagnostics.HasError() || strings.Contains(fmt.Sprint(resp.Diagnostics), "private-secret") {
				t.Fatal("unsafe destroy diagnostic", resp.Diagnostics)
			}
		})
	}
}
func TestSkillScanningInstanceRetainsNormalizedOrMissingRegistrationMetadata(t *testing.T) {
	created := false
	r, sr := skillResourceFixture(t, "instance", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			created = true
			_, _ = w.Write([]byte(`{"is_success":true,"tenant_id":"tenant","tsg_id":"123"}`))
			return
		}
		if !created {
			w.WriteHeader(404)
			return
		}
		_, _ = w.Write([]byte(`{"tenant_id":"tenant","tsg_id":"123","support_account_id":"normalized","created_by":"","support_account_name":null}`))
	})
	st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"tenant_id": types.StringValue("tenant"), "support_account_id": types.StringValue("support"), "created_by": types.StringValue("creator"), "support_account_name": types.StringValue("desired")}))
	resp := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: st.Raw}, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &resp)
	noSkillErrors(t, resp.Diagnostics)
	var obj types.Object
	noSkillErrors(t, resp.State.Get(context.Background(), &obj))
	if skillString(obj.Attributes(), "created_by") != "creator" || skillString(obj.Attributes(), "support_account_id") != "support" || skillString(obj.Attributes(), "support_account_name") != "desired" {
		t.Fatal("normalization changed planned inputs")
	}
	read := resource.ReadResponse{State: resp.State}
	r.Read(context.Background(), resource.ReadRequest{State: resp.State}, &read)
	noSkillErrors(t, read.Diagnostics)
	if !read.State.Raw.Equal(resp.State.Raw) {
		t.Fatal("normalized registration metadata caused permanent drift")
	}
}
func TestSkillScanningInstanceMalformedSuccessKeepsRequestedIdentity(t *testing.T) {
	r, sr := skillResourceFixture(t, "instance", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			_, _ = w.Write([]byte(`{"is_success":true,"tenant_id":"wrong"}`))
			return
		}
		w.WriteHeader(404)
	})
	st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"tenant_id": types.StringValue("tenant"), "support_account_id": types.StringValue("support"), "created_by": types.StringValue("creator")}))
	resp := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: st.Raw}, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &resp)
	if !resp.Diagnostics.HasError() || resp.State.Raw.IsNull() {
		t.Fatal("malformed success orphaned tenant identity")
	}
	var obj types.Object
	noSkillErrors(t, resp.State.Get(context.Background(), &obj))
	if skillString(obj.Attributes(), "id") != "tenant" {
		t.Fatal("wrong receipt identity adopted")
	}
}
func TestSkillScanningRuleCatalogDefaultAndFailedRestoration(t *testing.T) {
	for _, restoreFails := range []bool{false, true} {
		t.Run(fmt.Sprint(restoreFails), func(t *testing.T) {
			effective := ""
			r, sr := skillResourceFixture(t, "rule", func(w http.ResponseWriter, req *http.Request) {
				if req.Method == "PUT" {
					var b struct {
						Rules map[string]struct {
							State string `json:"state"`
						} `json:"rule_configurations"`
					}
					if e := json.NewDecoder(req.Body).Decode(&b); e != nil {
						t.Error(e)
					}
					if b.Rules[skillTestID].State == "ALLOWING" && restoreFails {
						_, _ = w.Write([]byte(`{"rule_instances":[],"pagination":{}}`))
						return
					}
					effective = b.Rules[skillTestID].State
				}
				if req.URL.Path == "/v1/rules" {
					if req.URL.Query().Get("skip") != "0" {
						_, _ = w.Write([]byte(`{"rules":[],"pagination":{}}`))
						return
					}
					_, _ = fmt.Fprintf(w, `{"rules":[{"uuid":%q,"name":"default","default_state":"ALLOWING"}],"pagination":{"total_items":1}}`, skillTestID)
					return
				}
				if effective == "" || req.URL.Query().Get("skip") != "0" {
					_, _ = w.Write([]byte(`{"rule_instances":[],"pagination":{}}`))
					return
				}
				_, _ = fmt.Fprintf(w, `{"rule_instances":[{"uuid":%q,"rule_uuid":%q,"tsg_id":"123","state":%q,"rule":{"name":"default"}}],"pagination":{}}`, skillTestID, skillTestID, effective)
			})
			st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"rule_uuid": types.StringValue(skillTestID), "state": types.StringValue("BLOCKING")}))
			create := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &create)
			noSkillErrors(t, create.Diagnostics)
			var obj types.Object
			noSkillErrors(t, create.State.Get(context.Background(), &obj))
			if skillString(obj.Attributes(), "original_state") != "ALLOWING" {
				t.Fatal("catalog default was not captured")
			}
			var destroy resource.DeleteResponse
			r.Delete(context.Background(), resource.DeleteRequest{State: create.State}, &destroy)
			if destroy.Diagnostics.HasError() != restoreFails {
				t.Fatal("failed restoration was accepted", destroy.Diagnostics)
			}
		})
	}
}
func TestSkillScanningOverrideRecoversMalformedCreateAndNullableTrust(t *testing.T) {
	for _, scenario := range []string{"missing uuid", "wrong tenant", "nullable trust"} {
		t.Run(scenario, func(t *testing.T) {
			created := false
			fingerprint := strings.Repeat("a", 64)
			row := map[string]any{"uuid": skillTestID, "tsg_id": "123", "skill_name": "skill", "fingerprint": fingerprint, "trusted_by": "trusted", "decision": "ALLOW", "created_at": "now", "updated_at": "now"}
			if scenario == "nullable trust" {
				row["trusted_by"] = nil
			}
			r, sr := skillResourceFixture(t, "override", func(w http.ResponseWriter, req *http.Request) {
				if req.Method == "POST" {
					created = true
					if scenario == "missing uuid" {
						_, _ = w.Write([]byte(`{"tsg_id":"123"}`))
						return
					}
					if scenario == "wrong tenant" {
						copy := map[string]any{"uuid": skillTestID, "tsg_id": "other"}
						_ = json.NewEncoder(w).Encode(copy)
						return
					}
					_ = json.NewEncoder(w).Encode(row)
					return
				}
				rows := []any{}
				if created && req.URL.Query().Get("skip") == "0" {
					rows = append(rows, row)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"skill_overrides": rows, "pagination": map[string]int{"total_items": len(rows)}})
			})
			st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"skill_name": types.StringValue("skill"), "fingerprint": types.StringValue(fingerprint), "trusted_by": types.StringValue("trusted")}))
			resp := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &resp)
			if resp.Diagnostics.HasError() != (scenario != "nullable trust") || resp.State.Raw.IsNull() {
				t.Fatal("create recovery result is incorrect", resp.Diagnostics)
			}
			var obj types.Object
			noSkillErrors(t, resp.State.Get(context.Background(), &obj))
			if skillString(obj.Attributes(), "id") != skillTestID {
				t.Fatal("identity not retained")
			}
			if scenario == "nullable trust" && skillString(obj.Attributes(), "trusted_by") != "trusted" {
				t.Fatal("nullable receipt lost required input")
			}
		})
	}
}

func TestSkillScanningInvalidEndpointsDoNotBlockModelSecurity(t *testing.T) {
	for _, key := range []string{"PANW_AGENT_GUARD_DATA_ENDPOINT", "PANW_AGENT_GUARD_MGMT_ENDPOINT"} {
		t.Setenv(key, "")
	}
	def := Definition()
	data, err := def.Configure(product.Credentials{ClientID: "id", ClientSecret: "secret", TsgID: "123"}, map[string]string{"skill_scanning_data_endpoint": "invalid-base"})
	if err != nil {
		t.Fatal("Skill Scanning configuration blocked unrelated products", err)
	}
	clients := data.(*Clients)
	if clients.Models == nil || clients.SkillsError == nil {
		t.Fatal("lazy Skill Scanning diagnostics were not retained")
	}
	slot := product.Data{"supply_chain": clients}
	if _, d := getModelSecClient(slot); d.HasError() {
		t.Fatal(d)
	}
	if _, d := getSkillClient(slot); !d.HasError() {
		t.Fatal("invalid endpoints accepted for Skill Scanning")
	}
}

func TestSkillScanningPopulatedFindingsAndChainDetailAreNative(t *testing.T) {
	for _, kind := range []string{"vulnerabilities", "attack_chains"} {
		t.Run(kind, func(t *testing.T) {
			r, _ := skillResourceFixture(t, "instance", func(w http.ResponseWriter, req *http.Request) {
				if kind == "vulnerabilities" {
					if req.URL.Query().Get("in_chain") != "false" || req.URL.Query().Get("type") != "SECRET_EXPOSURE" {
						t.Error("finding filters lost")
					}
					_, _ = fmt.Fprintf(w, `{"vulnerabilities":[{"uuid":%q,"tsg_id":"123","scan_uuid":%q,"title":"Synthetic secret","description":"fixture","file_path":null,"original_code":"synthetic-source","type_description":null,"vulnerability_type":"SECRET_EXPOSURE","attack_chain_count":1,"attack_chain_uuids":[%q]}],"pagination":{"total_items":1}}`, skillTestID, skillTestID, skillTestID)
					return
				}
				if !strings.HasSuffix(req.URL.Path, "/attack-chains/"+skillTestID) {
					t.Error("chain detail used list endpoint")
				}
				_, _ = fmt.Fprintf(w, `{"uuid":%q,"tsg_id":"123","scan_uuid":%q,"title":"Chain","description":"fixture","impact":"fixture","vulnerability_count":1,"vulnerability_uuids":[%q],"steps":[{"step_number":1,"step_type":"VULNERABILITY","node_id":"node","node_type":"TOOL","node_name":"tool","description":"step","impact_description":null}]}`, skillTestID, skillTestID, skillTestID)
			})
			d := &skillDataSource{kind: kind, clients: r.clients}
			var sr datasource.SchemaResponse
			d.Schema(context.Background(), datasource.SchemaRequest{}, &sr)
			ts, vs := map[string]attr.Type{}, map[string]attr.Value{}
			for k, a := range d.attributes() {
				ts[k] = a.GetType()
				v, e := ts[k].ValueFromTerraform(context.Background(), tftypes.NewValue(ts[k].TerraformType(context.Background()), nil))
				if e != nil {
					t.Fatal(e)
				}
				vs[k] = v
			}
			vs["scan_uuid"] = types.StringValue(skillTestID)
			if kind == "vulnerabilities" {
				vs["type"] = types.StringValue("SECRET_EXPOSURE")
				vs["in_chain"] = types.BoolValue(false)
			} else {
				vs["chain_uuid"] = types.StringValue(skillTestID)
			}
			st := tfsdk.State{Schema: sr.Schema}
			noSkillErrors(t, st.Set(context.Background(), types.ObjectValueMust(ts, vs)))
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: sr.Schema}}
			d.Read(context.Background(), datasource.ReadRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: st.Raw}}, &resp)
			noSkillErrors(t, resp.Diagnostics)
			var obj types.Object
			noSkillErrors(t, resp.State.Get(context.Background(), &obj))
			result := obj.Attributes()["result"].(types.Dynamic).UnderlyingValue().(types.Object).Attributes()
			if kind == "vulnerabilities" {
				row := result["vulnerabilities"].(types.Tuple).Elements()[0].(types.Object).Attributes()
				if !row["file_path"].IsNull() || row["original_code"].(types.String).ValueString() != "synthetic-source" {
					t.Fatal("finding fields lost")
				}
			} else {
				step := result["steps"].(types.Tuple).Elements()[0].(types.Object).Attributes()
				if step["node_type"].(types.String).ValueString() != "TOOL" || !step["impact_description"].IsNull() {
					t.Fatal("chain detail fields lost")
				}
			}
		})
	}
}

func TestSkillScanningFailedUpdateRetainsPriorDesiredState(t *testing.T) {
	r, sr := skillResourceFixture(t, "instance", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"message":"private-secret"}`))
	})
	prior := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"id": types.StringValue("tenant"), "tsg_id": types.StringValue("123"), "tenant_id": types.StringValue("tenant"), "support_account_id": types.StringValue("support"), "created_by": types.StringValue("original"), "iam_controlled": types.BoolValue(false), "auth_code_version": types.Int64Value(1)}))
	var object types.Object
	noSkillErrors(t, prior.Get(context.Background(), &object))
	values := object.Attributes()
	values["created_by"] = types.StringValue("changed")
	values["iam_controlled"] = types.BoolValue(true)
	values["auth_code_version"] = types.Int64Value(2)
	planned := skillState(t, sr, r.object(context.Background(), values))
	response := resource.UpdateResponse{State: planned}
	r.Update(context.Background(), resource.UpdateRequest{State: prior, Config: tfsdk.Config{Schema: sr.Schema, Raw: planned.Raw}, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: planned.Raw}}, &response)
	if !response.Diagnostics.HasError() || !response.State.Raw.Equal(prior.Raw) || strings.Contains(fmt.Sprint(response.Diagnostics), "private-secret") {
		t.Fatal("failed update recorded desired values as applied")
	}
}
func TestSkillScanningRegistrationRejectsJSONAndIdentityOverridesBeforeIOM(t *testing.T) {
	for _, payload := range []types.Dynamic{
		types.DynamicValue(types.StringValue(`{"region":"us"}`)),
		types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"tsg_id": types.StringType}, map[string]attr.Value{"tsg_id": types.StringValue("foreign")})),
		types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"region": types.StringType}, map[string]attr.Value{"region": types.StringUnknown()})),
	} {
		calls := 0
		r, sr := skillResourceFixture(t, "instance", func(http.ResponseWriter, *http.Request) { calls++ })
		st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"tenant_id": types.StringValue("tenant"), "created_by": types.StringValue("creator"), "support_account_id": types.StringValue("support"), "registration_details": payload}))
		response := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
		r.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: st.Raw}, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &response)
		if !response.Diagnostics.HasError() || calls != 0 {
			t.Fatal("invalid registration escaped to API")
		}
	}
}

func TestSkillScanningOverrideRetainsEmptyReasonOnServerNull(t *testing.T) {
	r := &skillResource{kind: "override"}
	v := map[string]attr.Value{"reason": types.StringValue(""), "trusted_by": types.StringValue("user")}
	r.mapOverride(v, &s.SkillOverrideResponse{})
	if v["reason"].IsNull() || skillString(v, "reason") != "" {
		t.Fatal("server null changed configured empty reason")
	}
	v["reason"] = types.StringNull()
	r.mapOverride(v, &s.SkillOverrideResponse{})
	if !v["reason"].IsNull() {
		t.Fatal("omitted reason must remain null")
	}
}

func TestSkillScanningForeignInstanceReceiptBlocksDelete(t *testing.T) {
	calls := 0
	r, sr := skillResourceFixture(t, "instance", func(w http.ResponseWriter, req *http.Request) {
		calls++
		if req.Method == "POST" {
			_, _ = w.Write([]byte(`{"is_success":true,"tenant_id":"tenant","tsg_id":"foreign"}`))
			return
		}
		w.WriteHeader(404)
	})
	st := skillState(t, sr, skillModel(t, r, map[string]attr.Value{"tenant_id": types.StringValue("tenant"), "support_account_id": types.StringValue("support"), "created_by": types.StringValue("user")}))
	create := resource.CreateResponse{State: tfsdk.State{Schema: sr.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: st.Raw}, Plan: tfsdk.Plan{Schema: sr.Schema, Raw: st.Raw}}, &create)
	if !create.Diagnostics.HasError() || create.State.Raw.IsNull() {
		t.Fatal("foreign receipt must checkpoint identity with error")
	}
	before := calls
	del := resource.DeleteResponse{State: create.State}
	r.Delete(context.Background(), resource.DeleteRequest{State: create.State}, &del)
	if !del.Diagnostics.HasError() || calls != before {
		t.Fatal("foreign receipt allowed tenant deletion")
	}
}
func TestSkillScanningRegistrationPreservesLargeNativeNumbers(t *testing.T) {
	n := json.Number("9007199254740993")
	raw := map[string]any{"entitlements": []any{map[string]any{"count": n}}, "tsg_instances": []any{map[string]any{"count": n}}, "extra": map[string]any{"count": n}, "extension": n}
	v, e := skillNative(context.Background(), raw)
	if e != nil {
		t.Fatal(e)
	}
	q, e := skillRegistration(types.DynamicValue(v))
	if e != nil {
		t.Fatal(e)
	}
	wire, e := json.Marshal(q)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Count(string(wire), string(n)) != 4 {
		t.Fatalf("registration lost numeric precision: %s", wire)
	}
}
func TestSkillScanningProductTokenEndpointEnvironmentIsolation(t *testing.T) {
	calls, hostileCalls := 0, 0
	hostile := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hostileCalls++; w.WriteHeader(400) }))
	defer hostile.Close()
	shared := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			calls++
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600}`))
			return
		}
		_, _ = w.Write([]byte(`{"rules":[]}`))
	}))
	defer shared.Close()
	t.Setenv("PANW_AGENT_GUARD_TOKEN_ENDPOINT", hostile.URL)
	t.Setenv("PANW_MODEL_SEC_TOKEN_ENDPOINT", shared.URL+"/token")
	ep := map[string]string{"mgmt_endpoint": shared.URL, "data_endpoint": shared.URL, "skill_scanning_mgmt_endpoint": shared.URL, "skill_scanning_data_endpoint": shared.URL}
	data, e := Definition().Configure(product.Credentials{ClientID: "id", ClientSecret: "secret", TsgID: "123", TokenEndpoint: shared.URL + "/token"}, ep)
	if e != nil {
		t.Fatal(e)
	}
	c := data.(*Clients)
	if _, e = c.Skills.Rules.List(context.Background(), ag.ListOpts{}); e != nil {
		t.Fatal(e)
	}
	// Existing Model Security environment behavior remains when shared override is absent.
	data, e = Definition().Configure(product.Credentials{ClientID: "id", ClientSecret: "secret", TsgID: "123"}, ep)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = data.(*Clients).Models.SecurityRules.List(context.Background(), modelsecurity.RuleListOpts{}); e != nil {
		t.Fatal(e)
	}
	if calls != 2 || hostileCalls != 0 {
		t.Fatal("token routing ignored shared options or changed Model Security fallback")
	}
}

func TestSkillScanningImportRejectsUppercaseUUIDBeforeIO(t *testing.T) {
	for _, kind := range []string{"rule", "override"} {
		calls := 0
		r, sr := skillResourceFixture(t, kind, func(http.ResponseWriter, *http.Request) { calls++ })
		resp := resource.ImportStateResponse{State: tfsdk.State{Schema: sr.Schema}}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: strings.ToUpper(skillTestID)}, &resp)
		if !resp.Diagnostics.HasError() || calls != 0 {
			t.Fatal("uppercase import reached API")
		}
	}
}
