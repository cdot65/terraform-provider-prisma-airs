package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const gatewayWorkspaceFixtureID = "67dc591a-9bdf-4460-ac93-ab6e3c1ae250"

type gatewayWorkspaceFixture struct {
	sync.Mutex
	scope, active                                          bool
	workspaceID, workspaceSlug                             string
	scopeDescription                                       string
	bindings                                               []any
	settings                                               map[string]any
	scopeCreates, workspaceCreates, archives, scopeDeletes int
}

func newGatewayWorkspaceFixture(t *testing.T) (*gatewayWorkspaceFixture, *httptest.Server) {
	t.Helper()
	f := &gatewayWorkspaceFixture{workspaceID: gatewayWorkspaceFixtureID, workspaceSlug: "stable-slug", settings: map[string]any{"name": "Application", "description": nil, "icon": nil, "defaults": nil, "usage_limits": nil, "rate_limits": nil}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		f.Lock()
		defer f.Unlock()
		w.Header().Set("Content-Type", "application/json")
		send := func(v any) {
			if err := json.NewEncoder(w).Encode(v); err != nil {
				t.Error(err)
			}
		}
		absent := func() { w.WriteHeader(404); send(map[string]any{"message": "absent"}) }
		body := func() map[string]any {
			v := map[string]any{}
			if err := json.NewDecoder(q.Body).Decode(&v); err != nil {
				t.Error(err)
			}
			return v
		}
		scope := func() map[string]any {
			bs := f.bindings
			if bs == nil {
				bs = []any{}
			}
			return map[string]any{"name": "tf_workspace_test", "description": f.scopeDescription, "resources": bs, "tsg_id": "123", "id": "tf_workspace_test:123"}
		}
		ws := func() map[string]any {
			b := map[string]any{"id": f.workspaceID, "slug": f.workspaceSlug, "scope_name": "tf_workspace_test", "created_at": "2026-10-03T00:00:00Z", "last_updated_at": "2026-10-03T00:00:00Z", "status": "active", "object": "workspace", "organisation_id": "123", "is_default": 0}
			for k, v := range f.settings {
				b[k] = v
			}
			return b
		}
		if q.URL.Path == "/token" {
			send(map[string]any{"access_token": "fixture", "expires_in": 3600})
			return
		}
		if q.Header.Get("Authorization") != "Bearer fixture" || q.Header.Get("x-tsg-id") != "123" {
			t.Error("workspace/IAM shared OAuth or tenant header missing")
		}
		switch q.URL.Path {
		case "/iam/scopes":
			b := body()
			f.scopeDescription, _ = b["description"].(string)
			f.scope = true
			f.scopeCreates++
			f.bindings = nil
			send(scope())
		case "/iam/scopes/tf_workspace_test":
			if !f.scope {
				absent()
				return
			}
			if q.Method == "DELETE" {
				f.scope = false
				f.scopeDeletes++
				send(map[string]any{})
				return
			}
			if q.Method == "PUT" {
				b := body()
				f.bindings, _ = b["resources"].([]any)
			}
			send(scope())
		case "/admin/workspaces":
			if q.Method == "POST" {
				b := body()
				for k, v := range b {
					if k != "scope_name" {
						f.settings[k] = v
					}
				}
				f.active = true
				f.workspaceCreates++
				if f.workspaceCreates > 1 {
					f.workspaceID = fmt.Sprintf("67dc591a-9bdf-4460-ac93-%012d", f.workspaceCreates)
					f.workspaceSlug = fmt.Sprintf("stable-slug-%d", f.workspaceCreates)
				}
				send(ws())
				return
			}
			rows := []any{}
			if f.active && q.URL.Query().Get("status") != "archived" {
				rows = append(rows, ws())
			}
			send(map[string]any{"object": "list", "data": rows, "total": len(rows), "has_more": false})
		case "/admin/workspaces/" + f.workspaceID:
			if !f.active {
				absent()
				return
			}
			if q.Method == "DELETE" {
				f.active = false
				f.archives++
				send(map[string]any{})
				return
			}
			if q.Method == "PUT" {
				b := body()
				for k, v := range b {
					f.settings[k] = v
				}
				send(map[string]any{})
				return
			}
			send(ws())
		default:
			t.Error("unexpected workspace lifecycle route", q.Method, q.URL.Path)
			w.WriteHeader(500)
		}
	}))
	t.Cleanup(server.Close)
	return f, server
}
func gatewayWorkspaceProviderConfig(server *httptest.Server) string {
	return fmt.Sprintf(`provider "prisma-airs" {
 client_id="fixture-client"
 client_secret="fixture-secret"
 tsg_id="123"
 token_endpoint=%q
 gateway {
  data_endpoint=%q
  admin_endpoint=%q
  iam_endpoint=%q
 }
}
`, server.URL+"/token", server.URL+"/data", server.URL+"/admin", server.URL+"/iam")
}
func TestGatewayWorkspaceTerraformNativeSettingsLifecycle(t *testing.T) {
	f, server := newGatewayWorkspaceFixture(t)
	base := gatewayWorkspaceProviderConfig(server)
	address := "prisma-airs_gateway_workspace.application"
	config := func(name string, limit int, settings bool) string {
		extras := ""
		if settings {
			extras = fmt.Sprintf(`icon="test"
 defaults={metadata={owner="terraform"}}
 usage_limits=[{type="tokens",credit_limit=%d}]
 rate_limits=[{type="requests",unit="rpm",value=60}]
`, limit)
		}
		return base + fmt.Sprintf(`resource "prisma-airs_gateway_workspace" "application" {
 name=%q
 scope_name="tf_workspace_test"
 scope_management="managed"
 %s
}
data "prisma-airs_gateway_workspace" "application" {workspace_id=prisma-airs_gateway_workspace.application.id}
data "prisma-airs_gateway_workspaces" "all" {}
`, name, extras)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("First label", 1000, true), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(address, "scope_owned", "true"), resource.TestCheckResourceAttr(address, "scope_binding_ready", "true"), resource.TestCheckResourceAttr(address, "usage_limits.0.credit_limit", "1000"), resource.TestCheckResourceAttr("data.prisma-airs_gateway_workspaces.all", "complete", "true"))},
		{Config: config("Updated label", 2000, true), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(address, "id", gatewayWorkspaceFixtureID), resource.TestCheckResourceAttr(address, "slug", "stable-slug"), resource.TestCheckResourceAttr(address, "usage_limits.0.credit_limit", "2000"))},
		{Config: config("Updated label", 2000, true), PlanOnly: true},
		{ResourceName: address, ImportState: true, ImportStateId: "managed/" + gatewayWorkspaceFixtureID + "/tf_workspace_test", ImportStateVerify: true, ImportStateVerifyIgnore: []string{"scope_ownership_token", "icon", "defaults", "usage_limits", "rate_limits"}},
		{Config: config("Updated label", 0, false), Check: func(*terraform.State) error {
			f.Lock()
			defer f.Unlock()
			d, _ := f.settings["defaults"].(map[string]any)
			metadata, _ := d["metadata"].(map[string]any)
			rates, _ := f.settings["rate_limits"].([]any)
			if len(metadata) != 0 || f.settings["usage_limits"] != nil || len(rates) != 0 || f.settings["icon"] != "" {
				return fmt.Errorf("configured settings were not cleared")
			}
			return nil
		}},
	}, CheckDestroy: func(*terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if f.scope || f.active || f.scopeCreates != 1 || f.workspaceCreates != 1 || f.archives != 1 || f.scopeDeletes != 1 {
			return fmt.Errorf("workspace lifecycle left fixtures or replaced stable identity")
		}
		return nil
	}})
}

func TestGatewayWorkspaceTerraformArchiveReplacementCleansScope(t *testing.T) {
	f, server := newGatewayWorkspaceFixture(t)
	addr := "prisma-airs_gateway_workspace.application"
	config := gatewayWorkspaceProviderConfig(server) + `resource "prisma-airs_gateway_workspace" "application" {
 name="Application"
 scope_name="tf_workspace_test"
 scope_management="managed"
}`
	var first string
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, ExpectNonEmptyPlan: true, Check: func(st *terraform.State) error {
			first = st.RootModule().Resources[addr].Primary.ID
			f.Lock()
			f.active = false
			f.Unlock()
			return nil
		}},
		{Config: config, Check: func(st *terraform.State) error {
			f.Lock()
			defer f.Unlock()
			if st.RootModule().Resources[addr].Primary.ID == first || f.scopeDeletes != 1 || f.workspaceCreates != 2 {
				return fmt.Errorf("archived workspace was not replaced after old owned scope cleanup")
			}
			return nil
		}},
		{Config: config, PlanOnly: true},
	}, CheckDestroy: func(*terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if f.scope || f.active || f.scopeDeletes != 2 {
			return fmt.Errorf("replacement workflow left a fixture")
		}
		return nil
	}})
}

func TestGatewayWorkspaceTerraformMissingExternalScopeNeverReplaces(t *testing.T) {
	f, server := newGatewayWorkspaceFixture(t)
	f.scope = true
	config := gatewayWorkspaceProviderConfig(server) + `resource "prisma-airs_gateway_workspace" "application" {
 name="Application"
 scope_name="tf_workspace_test"
 scope_management="external"
}`
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: resource.TestCheckResourceAttr("prisma-airs_gateway_workspace.application", "id", gatewayWorkspaceFixtureID)},
		{Config: config, PlanOnly: true, PreConfig: func() { f.Lock(); f.scope = false; f.Unlock() }, ExpectError: regexp.MustCompile("External IAM scope is missing")},
	}, CheckDestroy: func(*terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if f.active || f.workspaceCreates != 1 || f.archives != 1 || f.scopeCreates != 0 || f.scopeDeletes != 0 {
			return fmt.Errorf("missing external scope triggered replacement or IAM writes")
		}
		return nil
	}})
}
