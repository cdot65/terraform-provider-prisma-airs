package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWorkspaceManagedLifecyclePreservesLabelIdentityAndCleansScope(t *testing.T) {
	ctx := context.Background()
	scopeExists, bound, active := false, false, false
	name := "Initial label"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		reply := func(v any) {
			if e := json.NewEncoder(w).Encode(v); e != nil {
				t.Error(e)
			}
		}
		absent := func() { w.WriteHeader(http.StatusNotFound); reply(map[string]any{"message": "absent"}) }
		switch q.URL.Path {
		case "/token":
			reply(map[string]any{"access_token": "fixture", "expires_in": 3600})
		case "/iam/scopes":
			scopeExists = true
			reply(map[string]any{"name": "tf_apps", "description": "", "resources": []any{}, "tsg_id": "123", "id": "tf_apps:123"})
		case "/iam/scopes/tf_apps":
			if q.Method == http.MethodDelete {
				scopeExists = false
				reply(map[string]any{})
				return
			}
			if !scopeExists {
				absent()
				return
			}
			var rs []any
			if q.Method == http.MethodPut {
				bound = true
			}
			if bound {
				rs = []any{map[string]any{"resource_type": "workspace", "resource_id": "stable-slug", "metadata": []any{}}}
			} else {
				rs = []any{}
			}
			reply(map[string]any{"name": "tf_apps", "description": "", "resources": rs, "tsg_id": "123", "id": "tf_apps:123"})
		case "/admin/workspaces":
			if q.Method != http.MethodPost {
				reply(map[string]any{"data": []any{}, "total": 0, "has_more": false})
				return
			}
			if !scopeExists {
				t.Error("workspace created without IAM prerequisite")
			}
			active = true
			reply(map[string]any{"id": "workspace-uuid", "slug": "stable-slug", "name": name, "scope_name": "tf_apps", "description": nil, "created_at": "2026-10-03T00:00:00Z", "last_updated_at": "2026-10-03T00:00:00Z", "object": "workspace"})
		case "/admin/workspaces/workspace-uuid":
			if q.Method == http.MethodDelete {
				active = false
				reply(map[string]any{})
				return
			}
			if !active {
				absent()
				return
			}
			if q.Method == http.MethodPut {
				var b map[string]any
				if e := json.NewDecoder(q.Body).Decode(&b); e != nil {
					t.Error(e)
				}
				if n, ok := b["name"].(string); ok {
					name = n
				}
				reply(map[string]any{})
				return
			}
			reply(map[string]any{"id": "workspace-uuid", "slug": "stable-slug", "name": name, "scope_name": "tf_apps", "description": nil, "icon": nil, "defaults": nil, "usage_limits": nil, "rate_limits": nil, "created_at": "2026-10-03T00:00:00Z", "last_updated_at": "2026-10-03T00:00:00Z", "status": "active", "is_default": 0, "organisation_id": "123"})
		default:
			t.Error("unexpected route", q.Method, q.URL.Path)
			w.WriteHeader(500)
		}
	}))
	t.Cleanup(srv.Close)
	sdk, e := gw.NewClient(gw.Opts{ClientID: "client", ClientSecret: "secret", TsgID: "123", TokenEndpoint: srv.URL + "/token", DataEndpoint: srv.URL + "/data", AdminEndpoint: srv.URL + "/admin", IAMEndpoint: srv.URL + "/iam"})
	if e != nil {
		t.Fatal(e)
	}
	r := &workspaceResource{client: &client{sdk: sdk, organisation: "123"}}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	attrs := map[string]attr.Type{}
	values := map[string]attr.Value{}
	for k, a := range schema.Schema.Attributes {
		attrs[k] = a.GetType()
		v, e := nullNative(ctx, attrs[k])
		if e != nil {
			t.Fatal(e)
		}
		values[k] = v
	}
	values["name"] = types.StringValue(name)
	values["scope_name"] = types.StringValue("tf_apps")
	values["scope_management"] = types.StringValue("managed")
	planModel := types.ObjectValueMust(attrs, values)
	planState := tfsdk.State{Schema: schema.Schema}
	noErrors(t, planState.Set(ctx, planModel))
	create := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: schema.Schema, Raw: planState.Raw}}, &create)
	noErrors(t, create.Diagnostics)
	var current types.Object
	noErrors(t, create.State.Get(ctx, &current))
	if current.Attributes()["id"] != types.StringValue("workspace-uuid") || !bound {
		t.Fatal("workspace not provisioned with stable identity and scope binding")
	}
	values = current.Attributes()
	values["name"] = types.StringValue("Renamed label")
	planState = tfsdk.State{Schema: schema.Schema}
	noErrors(t, planState.Set(ctx, types.ObjectValueMust(attrs, values)))
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{State: create.State, Plan: tfsdk.Plan{Schema: schema.Schema, Raw: planState.Raw}}, &update)
	noErrors(t, update.Diagnostics)
	noErrors(t, update.State.Get(ctx, &current))
	if current.Attributes()["id"] != types.StringValue("workspace-uuid") || current.Attributes()["slug"] != types.StringValue("stable-slug") || name != "Renamed label" {
		t.Fatal("renaming replaced workspace identity")
	}
	destroy := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &destroy)
	noErrors(t, destroy.Diagnostics)
	if scopeExists || active {
		t.Fatalf("owned fixtures remain: scope=%t active=%t", scopeExists, active)
	}
}

type workspaceHarness struct {
	mu                                           sync.Mutex
	scope, active, everCreated                   bool
	resources                                    []gw.IAMScopeResource
	fields                                       document
	iamWrites, workspaceWrites                   int
	ambiguousCreate                              bool
	failBinding, failScopeDelete, ambiguousScope bool
	scopeDescription                             string
	omitHasMore                                  bool
	decorateName                                 bool
	ignoreClear                                  bool
	inventoryExtra                               int
}

func newWorkspaceHarness(t *testing.T, external bool) (*workspaceResource, resource.SchemaResponse, *workspaceHarness) {
	t.Helper()
	h := &workspaceHarness{scope: external, fields: document{"name": "Application", "description": nil, "icon": nil, "defaults": nil, "usage_limits": nil, "rate_limits": nil}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		h.mu.Lock()
		defer h.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		reply := func(v any) {
			if e := json.NewEncoder(w).Encode(v); e != nil {
				t.Error(e)
			}
		}
		notFound := func() { w.WriteHeader(404); reply(document{"message": "absent"}) }
		if q.URL.Path == "/token" {
			reply(document{"access_token": "fixture", "expires_in": 3600})
			return
		}
		if strings.HasPrefix(q.URL.Path, "/iam/") && q.Method != "GET" {
			h.iamWrites++
		}
		if strings.HasPrefix(q.URL.Path, "/admin/") && q.Method != "GET" {
			h.workspaceWrites++
		}
		switch q.URL.Path {
		case "/iam/scopes":
			var input document
			if e := json.NewDecoder(q.Body).Decode(&input); e != nil {
				t.Error(e)
			}
			h.scopeDescription, _ = input["description"].(string)
			h.scope = true
			if h.ambiguousScope {
				w.WriteHeader(503)
				reply(document{"message": "ambiguous"})
				return
			}
			reply(document{"name": "tf_apps", "description": h.scopeDescription, "resources": []any{}, "tsg_id": "123", "id": "tf_apps:123"})
		case "/iam/scopes/tf_apps":
			if !h.scope {
				notFound()
				return
			}
			if q.Method == "DELETE" {
				if h.failScopeDelete {
					w.WriteHeader(503)
					reply(document{"message": "cleanup blocked"})
					return
				}
				h.scope = false
				reply(document{})
				return
			}
			if q.Method == "PUT" {
				if h.failBinding {
					w.WriteHeader(503)
					reply(document{"message": "bind blocked"})
					return
				}
				var b gw.IAMScope
				if e := json.NewDecoder(q.Body).Decode(&b); e != nil {
					t.Error(e)
				}
				h.resources = b.Resources
			}
			rs := h.resources
			if rs == nil {
				rs = []gw.IAMScopeResource{}
			}
			reply(document{"name": "tf_apps", "description": h.scopeDescription, "resources": rs, "tsg_id": "123", "id": "tf_apps:123"})
		case "/admin/workspaces":
			if q.Method == "POST" {
				var b document
				if e := json.NewDecoder(q.Body).Decode(&b); e != nil {
					t.Error(e)
				}
				for k, v := range b {
					if k != "scope_name" {
						h.fields[k] = v
					}
				}
				h.active = true
				h.everCreated = true
				if h.ambiguousCreate {
					w.WriteHeader(503)
					reply(document{"message": "ambiguous"})
					return
				}
				reply(document{"id": "workspace-uuid", "slug": "stable-slug", "name": h.fields["name"], "description": h.fields["description"], "scope_name": "tf_apps", "created_at": "2026-10-03T00:00:00Z", "last_updated_at": "2026-10-03T00:00:00Z", "object": "workspace"})
				return
			}
			rows := []any{}
			if (h.active && q.URL.Query().Get("status") != "archived") || (h.everCreated && !h.active && q.URL.Query().Get("status") == "archived") {
				status := "active"
				if !h.active {
					status = "archived"
				}
				rows = append(rows, document{"id": "workspace-uuid", "slug": "stable-slug", "name": h.fields["name"], "description": h.fields["description"], "scope_name": "tf_apps", "status": status, "icon": h.fields["icon"], "object": "workspace", "is_default": 0, "created_at": "2026-10-03T00:00:00Z", "last_updated_at": "2026-10-03T00:00:00Z"})
			}
			envelope := document{"object": "list", "data": rows, "total": len(rows) + h.inventoryExtra}
			if !h.omitHasMore {
				envelope["has_more"] = false
			}
			reply(envelope)
		case "/admin/workspaces/workspace-uuid":
			if !h.active {
				notFound()
				return
			}
			if q.Method == "DELETE" {
				h.active = false
				reply(document{})
				return
			}
			if q.Method == "PUT" {
				var b document
				if e := json.NewDecoder(q.Body).Decode(&b); e != nil {
					t.Error(e)
				}
				if !h.ignoreClear {
					for k, v := range b {
						h.fields[k] = v
					}
				}
				reply(document{})
				return
			}
			b := document{"id": "workspace-uuid", "slug": "stable-slug", "scope_name": "tf_apps", "created_at": "2026-10-03T00:00:00Z", "last_updated_at": "2026-10-03T00:00:00Z", "is_default": 0}
			for k, v := range h.fields {
				b[k] = v
			}
			if h.decorateName {
				if icon, ok := b["icon"].(string); ok && icon != "" {
					b["name"] = icon + " " + h.fields["name"].(string)
				}
			}
			reply(b)
		default:
			t.Error("unexpected workspace contract route", q.URL.Path)
			w.WriteHeader(500)
		}
	}))
	t.Cleanup(server.Close)
	sdk, e := gw.NewClient(gw.Opts{ClientID: "client", ClientSecret: "secret", TsgID: "123", TokenEndpoint: server.URL + "/token", DataEndpoint: server.URL + "/data", AdminEndpoint: server.URL + "/admin", IAMEndpoint: server.URL + "/iam"})
	if e != nil {
		t.Fatal(e)
	}
	r := &workspaceResource{client: &client{sdk: sdk, organisation: "123"}}
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	return r, s, h
}
func workspaceModel(t *testing.T, s resource.SchemaResponse, mode string, extra map[string]attr.Value) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	ts := map[string]attr.Type{}
	vs := map[string]attr.Value{}
	for k, a := range s.Schema.Attributes {
		ts[k] = a.GetType()
		v, e := nullNative(ctx, ts[k])
		if e != nil {
			t.Fatal(e)
		}
		vs[k] = v
	}
	vs["name"] = types.StringValue("Application")
	vs["scope_name"] = types.StringValue("tf_apps")
	vs["scope_management"] = types.StringValue(mode)
	for k, v := range extra {
		vs[k] = v
	}
	st := tfsdk.State{Schema: s.Schema}
	noErrors(t, st.Set(ctx, types.ObjectValueMust(ts, vs)))
	return st
}
func workspaceCreate(t *testing.T, r *workspaceResource, s resource.SchemaResponse, mode string) tfsdk.State {
	t.Helper()
	st := workspaceModel(t, s, mode, nil)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &resp)
	noErrors(t, resp.Diagnostics)
	return resp.State
}
func TestWorkspaceExternalScopeReceivesNoWrites(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, true)
	h.resources = []gw.IAMScopeResource{{ResourceType: "other", ResourceID: "external-object", Metadata: []any{}}}
	state := workspaceCreate(t, r, s, "external")
	resp := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)
	noErrors(t, resp.Diagnostics)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.iamWrites != 0 || !h.scope || len(h.resources) != 1 || h.resources[0].ResourceID != "external-object" {
		t.Fatal("external IAM scope was modified")
	}
}
func TestWorkspaceDestroyRefusesNewlySharedOwnedScope(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	state := workspaceCreate(t, r, s, "managed")
	h.mu.Lock()
	h.resources = append(h.resources, gw.IAMScopeResource{ResourceType: "workspace", ResourceID: "other-workspace", Metadata: []any{}})
	before := h.iamWrites
	h.mu.Unlock()
	resp := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("destroy accepted a newly shared scope")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.active || !h.scope || h.iamWrites != before {
		t.Fatal("destruction affected a shared scope or workspace")
	}
}

func TestWorkspaceImportRequiresExplicitDedicatedScopeAdoption(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, true)
	h.active = true
	h.resources = []gw.IAMScopeResource{{ResourceType: "workspace", ResourceID: "stable-slug", Metadata: []any{}}}
	resp := resource.ImportStateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "workspace-uuid/tf_apps"}, &resp)
	noErrors(t, resp.Diagnostics)
	var m types.Object
	noErrors(t, resp.State.Get(context.Background(), &m))
	if workspaceValue(m, "scope_management") != "external" || workspaceFlag(m, "scope_owned") {
		t.Fatal("import automatically assumed IAM ownership")
	}
	owned := resource.ImportStateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "managed/workspace-uuid/tf_apps"}, &owned)
	noErrors(t, owned.Diagnostics)
	noErrors(t, owned.State.Get(context.Background(), &m))
	if !workspaceFlag(m, "scope_owned") {
		t.Fatal("explicit dedicated scope adoption was not recorded")
	}
	h.resources = append(h.resources, gw.IAMScopeResource{ResourceType: "workspace", ResourceID: "other-workspace", Metadata: []any{}})
	unsafe := resource.ImportStateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "managed/workspace-uuid/tf_apps"}, &unsafe)
	if !unsafe.Diagnostics.HasError() {
		t.Fatal("shared scope adoption accepted")
	}
}

func TestWorkspaceRefreshRetainsOwnedScopeForArchiveReplacement(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	state := workspaceCreate(t, r, s, "managed")
	h.mu.Lock()
	h.active = false
	before := h.iamWrites
	h.mu.Unlock()
	read := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &read)
	noErrors(t, read.Diagnostics)
	var m types.Object
	noErrors(t, read.State.Get(context.Background(), &m))
	if workspaceValue(m, "status") != "archived" || !workspaceFlag(m, "scope_owned") {
		t.Fatal("refresh forgot an archived workspace's owned IAM scope")
	}
	planned := workspaceModel(t, s, "managed", nil)
	change := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: s.Schema, Raw: planned.Raw}}
	r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{State: read.State, Plan: change.Plan, Config: tfsdk.Config{Schema: s.Schema, Raw: planned.Raw}}, &change)
	noErrors(t, change.Diagnostics)
	if len(change.RequiresReplace) == 0 {
		t.Fatal("external archive did not plan replacement")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.scope || h.iamWrites != before {
		t.Fatal("refresh or planning mutated IAM")
	}
}

func TestWorkspaceAmbiguousCreateRecoversIdentityBeforeScopeCleanup(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	h.ambiguousCreate = true
	st := workspaceModel(t, s, "managed", nil)
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	if !created.Diagnostics.HasError() {
		t.Fatal("ambiguous create reported success")
	}
	var m types.Object
	noErrors(t, created.State.Get(context.Background(), &m))
	if !workspaceFlag(m, "scope_owned") {
		t.Fatal("acknowledged scope identity was lost")
	}
	destroy := resource.DeleteResponse{State: created.State}
	r.Delete(context.Background(), resource.DeleteRequest{State: created.State}, &destroy)
	noErrors(t, destroy.Diagnostics)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.active || h.scope {
		t.Fatal("scope cleanup left an ambiguous-created active workspace")
	}
}

func TestWorkspaceDiscoveryStoresOnlySafeMetadata(t *testing.T) {
	r, _, h := newWorkspaceHarness(t, true)
	h.active = true
	h.fields["defaults"] = document{"metadata": document{"openai_api_key": "never-store-this-credential"}}
	d := &workspaceDataSource{client: r.client}
	ctx := context.Background()
	var s datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &s)
	ts := map[string]attr.Type{}
	values := map[string]attr.Value{}
	for k, a := range s.Schema.Attributes {
		ts[k] = a.GetType()
		values[k], _ = nullNative(ctx, ts[k])
	}
	values["workspace_id"] = types.StringValue("workspace-uuid")
	cfg := tfsdk.State{Schema: s.Schema}
	noErrors(t, cfg.Set(ctx, types.ObjectValueMust(ts, values)))
	read := datasource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: s.Schema, Raw: cfg.Raw}}, &read)
	noErrors(t, read.Diagnostics)
	var result types.Object
	noErrors(t, read.State.Get(ctx, &result))
	if result.Attributes()["scope_name"] != types.StringValue("tf_apps") {
		t.Fatal("scope metadata lost")
	}
	if strings.Contains(fmt.Sprint(result), "never-store-this-credential") {
		t.Fatal("discovery leaked a configuration credential")
	}
}

func TestWorkspaceResumesBindingWithoutRepeatingCreate(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	h.failBinding = true
	st := workspaceModel(t, s, "managed", nil)
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	if !created.Diagnostics.HasError() {
		t.Fatal("binding failure reported success")
	}
	var old types.Object
	noErrors(t, created.State.Get(context.Background(), &old))
	if workspaceValue(old, "id") != "workspace-uuid" || workspaceValue(old, "provisioning_stage") != "workspace_created" {
		t.Fatal("partial identity was lost")
	}
	h.mu.Lock()
	h.failBinding = false
	before := h.workspaceWrites
	h.mu.Unlock()
	updated := resource.UpdateResponse{State: created.State}
	r.Update(context.Background(), resource.UpdateRequest{State: created.State, Plan: tfsdk.Plan{Schema: s.Schema, Raw: created.State.Raw}}, &updated)
	noErrors(t, updated.Diagnostics)
	var after types.Object
	noErrors(t, updated.State.Get(context.Background(), &after))
	if !workspaceFlag(after, "scope_binding_ready") || workspaceValue(after, "provisioning_stage") != "ready" {
		t.Fatal("binding not resumed")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.workspaceWrites != before || len(h.resources) != 1 {
		t.Fatal("resuming provisioning repeated workspace creation or missed binding")
	}
}
func TestWorkspaceFailedScopeCleanupRetainsArchivedCheckpoint(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	state := workspaceCreate(t, r, s, "managed")
	h.failScopeDelete = true
	destroyed := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &destroyed)
	if !destroyed.Diagnostics.HasError() {
		t.Fatal("cleanup failure reported success")
	}
	var checkpoint types.Object
	noErrors(t, destroyed.State.Get(context.Background(), &checkpoint))
	if workspaceValue(checkpoint, "provisioning_stage") != "archived" {
		t.Fatal("archival was not checkpointed before failed scope deletion")
	}
	h.mu.Lock()
	h.failScopeDelete = false
	h.mu.Unlock()
	retried := resource.DeleteResponse{State: destroyed.State}
	r.Delete(context.Background(), resource.DeleteRequest{State: destroyed.State}, &retried)
	noErrors(t, retried.Diagnostics)
	if !retried.State.Raw.IsNull() {
		t.Fatal("cleanup retry did not remove state")
	}
}
func TestWorkspaceAmbiguousScopeCreationRetainsRecoveryIntent(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	h.ambiguousScope = true
	st := workspaceModel(t, s, "managed", nil)
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	if !created.Diagnostics.HasError() {
		t.Fatal("ambiguous scope creation reported success")
	}
	if created.State.Raw.IsNull() {
		t.Fatal("scope creation intent lost")
	}
	var m types.Object
	noErrors(t, created.State.Get(context.Background(), &m))
	if workspaceValue(m, "provisioning_stage") != "scope_create_uncertain" || workspaceValue(m, "scope_ownership_token") == "" {
		t.Fatal("scope recovery intent missing")
	}
	read := resource.ReadResponse{State: created.State}
	r.Read(context.Background(), resource.ReadRequest{State: created.State}, &read)
	noErrors(t, read.Diagnostics)
	noErrors(t, read.State.Get(context.Background(), &m))
	if !workspaceFlag(m, "scope_owned") || workspaceValue(m, "provisioning_stage") != "scope_ready" {
		t.Fatal("matching scope intent not reconciled")
	}
	destroyed := resource.DeleteResponse{State: read.State}
	r.Delete(context.Background(), resource.DeleteRequest{State: read.State}, &destroyed)
	noErrors(t, destroyed.Diagnostics)
	if h.scope || h.active {
		t.Fatal("ambiguous scope cleanup left owned fixtures")
	}
}

func TestWorkspaceInventoryDoesNotInventPaginationFlag(t *testing.T) {
	for _, tc := range []struct {
		name    string
		omit    bool
		extra   int
		hasNull bool
	}{{"flag omitted", true, 0, true}, {"inconsistent total", false, 1, false}} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, h := newWorkspaceHarness(t, true)
			h.omitHasMore = tc.omit
			h.inventoryExtra = tc.extra
			d := &workspaceDataSource{client: r.client, list: true}
			ctx := context.Background()
			var s datasource.SchemaResponse
			d.Schema(ctx, datasource.SchemaRequest{}, &s)
			ts := map[string]attr.Type{}
			vs := map[string]attr.Value{}
			for k, a := range s.Schema.Attributes {
				ts[k] = a.GetType()
				vs[k], _ = nullNative(ctx, ts[k])
			}
			st := tfsdk.State{Schema: s.Schema}
			noErrors(t, st.Set(ctx, types.ObjectValueMust(ts, vs)))
			got := datasource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
			d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: s.Schema, Raw: st.Raw}}, &got)
			noErrors(t, got.Diagnostics)
			var m types.Object
			noErrors(t, got.State.Get(ctx, &m))
			if m.Attributes()["has_more"].IsNull() != tc.hasNull || (!tc.hasNull && workspaceFlag(m, "has_more")) || workspaceFlag(m, "complete") {
				t.Fatal("incomplete inventory fabricated a pagination flag or completeness")
			}
		})
	}
}

func TestWorkspaceSettingsUseTypedNativeHCLAndRefreshOwnedDrift(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	defaultType, ok := s.Schema.Attributes["defaults"].GetType().(types.ObjectType)
	if !ok {
		t.Fatal("defaults should expose native named HCL fields")
	}
	defaults := types.ObjectValueMust(defaultType.AttrTypes, map[string]attr.Value{"config_id": types.StringNull(), "metadata": types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"owner": types.StringType}, map[string]attr.Value{"owner": types.StringValue("terraform")}))})
	st := workspaceModel(t, s, "managed", map[string]attr.Value{"defaults": defaults})
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	noErrors(t, created.Diagnostics)
	h.mu.Lock()
	h.fields["defaults"] = map[string]any{"config_id": nil, "metadata": map[string]any{"owner": "external-change"}}
	h.mu.Unlock()
	read := resource.ReadResponse{State: created.State}
	r.Read(context.Background(), resource.ReadRequest{State: created.State}, &read)
	noErrors(t, read.Diagnostics)
	var actual types.Object
	noErrors(t, read.State.Get(context.Background(), &actual))
	value, e := nativeJSON(actual.Attributes()["defaults"])
	if e != nil {
		t.Fatal(e)
	}
	if value.(map[string]any)["metadata"].(map[string]any)["owner"] != "external-change" {
		t.Fatal("managed default drift hidden by refresh")
	}
}

func TestWorkspaceRemovingSettingsClearsOnlyPreviouslyOwnedFields(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	ctx := context.Background()
	defaultsType := s.Schema.Attributes["defaults"].GetType().(types.ObjectType)
	defaultValue := types.ObjectValueMust(defaultsType.AttrTypes, map[string]attr.Value{"config_id": types.StringNull(), "metadata": types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{"owner": types.StringType}, map[string]attr.Value{"owner": types.StringValue("terraform")}))})
	st := workspaceModel(t, s, "managed", map[string]attr.Value{"defaults": defaultValue, "icon": types.StringValue("populated")})
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	noErrors(t, created.Diagnostics)
	h.mu.Lock()
	h.fields["rate_limits"] = []any{map[string]any{"type": "requests", "unit": "rpm", "value": 60}}
	h.mu.Unlock()
	removed := workspaceModel(t, s, "managed", nil)
	updated := resource.UpdateResponse{State: created.State}
	r.Update(ctx, resource.UpdateRequest{State: created.State, Plan: tfsdk.Plan{Schema: s.Schema, Raw: removed.Raw}}, &updated)
	noErrors(t, updated.Diagnostics)
	h.mu.Lock()
	defer h.mu.Unlock()
	d := h.fields["defaults"].(map[string]any)
	if len(d["metadata"].(map[string]any)) != 0 || h.fields["icon"] != "" {
		t.Fatal("previously managed settings were not cleared")
	}
	if len(h.fields["rate_limits"].([]any)) != 1 {
		t.Fatal("never-managed rate policy was cleared")
	}
}

func TestWorkspaceIgnoredClearingRetainsPriorState(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	st := workspaceModel(t, s, "managed", map[string]attr.Value{"icon": types.StringValue("populated")})
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	noErrors(t, created.Diagnostics)
	h.mu.Lock()
	h.ignoreClear = true
	h.mu.Unlock()
	removed := workspaceModel(t, s, "managed", nil)
	updated := resource.UpdateResponse{State: created.State}
	r.Update(context.Background(), resource.UpdateRequest{State: created.State, Plan: tfsdk.Plan{Schema: s.Schema, Raw: removed.Raw}}, &updated)
	if !updated.Diagnostics.HasError() {
		t.Fatal("acknowledgement without actual clearing reported success")
	}
	var m types.Object
	noErrors(t, updated.State.Get(context.Background(), &m))
	if workspaceValue(m, "icon") != "populated" {
		t.Fatal("failed clearing discarded the owned prior value")
	}
}

func TestWorkspaceRefusesScopeCollisionAndForeignRecoveryToken(t *testing.T) {
	t.Run("existing name", func(t *testing.T) {
		r, s, h := newWorkspaceHarness(t, true)
		st := workspaceModel(t, s, "managed", nil)
		got := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
		r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &got)
		if !got.Diagnostics.HasError() || h.iamWrites != 0 || h.workspaceWrites != 0 {
			t.Fatal("existing scope was adopted or modified")
		}
	})
	t.Run("uncertain marker mismatch", func(t *testing.T) {
		r, s, h := newWorkspaceHarness(t, false)
		h.ambiguousScope = true
		st := workspaceModel(t, s, "managed", nil)
		got := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
		r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &got)
		if !got.Diagnostics.HasError() {
			t.Fatal("missing ambiguous failure")
		}
		h.mu.Lock()
		h.scopeDescription = "foreign scope"
		before := h.iamWrites
		h.mu.Unlock()
		destroy := resource.DeleteResponse{State: got.State}
		r.Delete(context.Background(), resource.DeleteRequest{State: got.State}, &destroy)
		if !destroy.Diagnostics.HasError() || h.iamWrites != before || !h.scope {
			t.Fatal("mismatched intent acquired ownership")
		}
	})
}
func TestWorkspaceIncompleteInventoryCannotAuthorizeAmbiguousCleanup(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	h.ambiguousCreate = true
	h.omitHasMore = true
	st := workspaceModel(t, s, "managed", nil)
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	if !created.Diagnostics.HasError() {
		t.Fatal("ambiguous creation reported success")
	}
	h.mu.Lock()
	beforeIAM, beforeWorkspace := h.iamWrites, h.workspaceWrites
	h.mu.Unlock()
	destroy := resource.DeleteResponse{State: created.State}
	r.Delete(context.Background(), resource.DeleteRequest{State: created.State}, &destroy)
	if !destroy.Diagnostics.HasError() {
		t.Fatal("incomplete inventory authorised cleanup")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.scope || !h.active || beforeIAM != h.iamWrites || beforeWorkspace != h.workspaceWrites {
		t.Fatal("ambiguous objects modified without conclusive inventory")
	}
}
func TestWorkspaceBindingResumeRefusesNewSharedScope(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	h.failBinding = true
	st := workspaceModel(t, s, "managed", nil)
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	if !created.Diagnostics.HasError() {
		t.Fatal("binding unexpectedly succeeded")
	}
	h.mu.Lock()
	h.failBinding = false
	h.resources = []gw.IAMScopeResource{{ResourceType: "workspace", ResourceID: "foreign-slug", Metadata: []any{}}}
	before := h.iamWrites
	h.mu.Unlock()
	update := resource.UpdateResponse{State: created.State}
	r.Update(context.Background(), resource.UpdateRequest{State: created.State, Plan: tfsdk.Plan{Schema: s.Schema, Raw: created.State.Raw}}, &update)
	if !update.Diagnostics.HasError() {
		t.Fatal("newly shared scope accepted")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.iamWrites != before || len(h.resources) != 1 || h.resources[0].ResourceID != "foreign-slug" {
		t.Fatal("shared binding overwritten")
	}
}

func TestWorkspaceIconDecorationPreservesLogicalLabel(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	h.decorateName = true
	st := workspaceModel(t, s, "managed", map[string]attr.Value{"icon": types.StringValue("test")})
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	noErrors(t, created.Diagnostics)
	read := resource.ReadResponse{State: created.State}
	r.Read(context.Background(), resource.ReadRequest{State: created.State}, &read)
	noErrors(t, read.Diagnostics)
	var m types.Object
	noErrors(t, read.State.Get(context.Background(), &m))
	if workspaceValue(m, "name") != "Application" {
		t.Fatal("display icon became part of logical Terraform label")
	}
}

func TestWorkspaceBindingResumeAppliesChangedPlanAfterRecovery(t *testing.T) {
	r, s, h := newWorkspaceHarness(t, false)
	h.failBinding = true
	st := workspaceModel(t, s, "managed", nil)
	created := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &created)
	if !created.Diagnostics.HasError() {
		t.Fatal("binding unexpectedly succeeded")
	}
	h.mu.Lock()
	h.failBinding = false
	h.mu.Unlock()
	var m types.Object
	noErrors(t, created.State.Get(context.Background(), &m))
	vs := m.Attributes()
	vs["name"] = types.StringValue("Changed recovery label")
	vs["status"] = types.StringUnknown()
	vs["created_at"] = types.StringUnknown()
	vs["last_updated_at"] = types.StringUnknown()
	planned := tfsdk.State{Schema: s.Schema}
	noErrors(t, planned.Set(context.Background(), types.ObjectValueMust(m.AttributeTypes(context.Background()), vs)))
	updated := resource.UpdateResponse{State: created.State}
	r.Update(context.Background(), resource.UpdateRequest{State: created.State, Plan: tfsdk.Plan{Schema: s.Schema, Raw: planned.Raw}}, &updated)
	noErrors(t, updated.Diagnostics)
	noErrors(t, updated.State.Get(context.Background(), &m))
	if workspaceValue(m, "name") != "Changed recovery label" || workspaceValue(m, "id") != "workspace-uuid" || workspaceValue(m, "provisioning_stage") != "ready" {
		t.Fatal("recovery did not apply changed configuration at stable identity")
	}
}

func TestWorkspaceMissingScopePreservesServiceStatusAndExternalWorkspace(t *testing.T) {
	for _, mode := range []string{"managed", "external"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			r, s, h := newWorkspaceHarness(t, mode == "external")
			state := workspaceCreate(t, r, s, mode)
			h.mu.Lock()
			h.scope = false
			iamBefore, wsBefore := h.iamWrites, h.workspaceWrites
			h.mu.Unlock()
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &read)
			noErrors(t, read.Diagnostics)
			var m types.Object
			noErrors(t, read.State.Get(ctx, &m))
			if workspaceValue(m, "status") == "scope_missing" || workspaceValue(m, "id") != "workspace-uuid" || workspaceFlag(m, "scope_binding_ready") {
				t.Fatal("missing scope corrupted service status or workspace identity/binding")
			}
			planned := workspaceModel(t, s, mode, nil)
			change := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: s.Schema, Raw: planned.Raw}}
			r.ModifyPlan(ctx, resource.ModifyPlanRequest{State: read.State, Plan: change.Plan}, &change)
			if change.Diagnostics.HasError() != (mode == "external") {
				t.Fatalf("unexpected missing-scope planning diagnostics: %v", change.Diagnostics)
			}
			if (len(change.RequiresReplace) > 0) != (mode == "managed") {
				t.Fatal("replacement did not respect scope ownership")
			}
			h.mu.Lock()
			defer h.mu.Unlock()
			if !h.active || h.iamWrites != iamBefore || h.workspaceWrites != wsBefore {
				t.Fatal("refresh/planning mutated remote objects")
			}
		})
	}
}
func TestWorkspaceDiscoveryPreservesLiteralIconPrefix(t *testing.T) {
	for _, plural := range []bool{false, true} {
		t.Run(fmt.Sprint(plural), func(t *testing.T) {
			ctx := context.Background()
			r, _, h := newWorkspaceHarness(t, true)
			h.active, h.everCreated, h.decorateName = true, true, true
			h.fields["name"], h.fields["icon"] = "test Application", "test"
			d := &workspaceDataSource{client: r.client, list: plural}
			var s datasource.SchemaResponse
			d.Schema(ctx, datasource.SchemaRequest{}, &s)
			ts := map[string]attr.Type{}
			vs := map[string]attr.Value{}
			for k, a := range s.Schema.Attributes {
				ts[k] = a.GetType()
				vs[k], _ = nullNative(ctx, ts[k])
			}
			if !plural {
				vs["workspace_id"] = types.StringValue("workspace-uuid")
			}
			cfg := tfsdk.State{Schema: s.Schema}
			noErrors(t, cfg.Set(ctx, types.ObjectValueMust(ts, vs)))
			read := datasource.ReadResponse{State: tfsdk.State{Schema: s.Schema}}
			d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: s.Schema, Raw: cfg.Raw}}, &read)
			noErrors(t, read.Diagnostics)
			var got types.Object
			noErrors(t, read.State.Get(ctx, &got))
			label := got.Attributes()["name"]
			if plural {
				label = got.Attributes()["items"].(types.List).Elements()[0].(types.Object).Attributes()["name"]
			}
			if label != types.StringValue("test Application") {
				t.Fatalf("literal icon prefix lost: %v", label)
			}
		})
	}
}
