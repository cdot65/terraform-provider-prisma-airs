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
)

func TestAdapterMockLifecycleAndReadOnlyImport(t *testing.T) {
	var lock sync.Mutex
	var saved map[string]any
	var storedSecret string
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		send := func(v any) {
			if err := json.NewEncoder(w).Encode(v); err != nil {
				t.Error(err)
			}
		}
		if q.URL.Path == "/token" {
			send(map[string]any{"access_token": "fixture", "expires_in": 3600})
			return
		}
		if q.URL.Path != "/v1/adapters" && q.URL.Path != "/v1/adapters/fixture-adapter" {
			t.Errorf("unexpected path %s", q.URL.Path)
			w.WriteHeader(404)
			return
		}
		if q.Method == "DELETE" {
			saved = nil
			send(map[string]any{})
			return
		}
		if q.Method == "POST" || q.Method == "PUT" {
			if q.URL.Query().Get("validate") != "false" {
				t.Error("draft write did not explicitly disable execution")
			}
			var body map[string]any
			if err := json.NewDecoder(q.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			for _, item := range body["variables"].([]any) {
				v := item.(map[string]any)
				if v["key"] == "credential" {
					if v["value"] != nil {
						storedSecret = v["value"].(string)
					} else if storedSecret == "" {
						t.Error("null attempted to initialize secret")
					}
				}
			}
			saved = body
			writes++
		}
		if q.URL.Path == "/v1/adapters" && q.Method == "GET" {
			send(map[string]any{"data": []any{map[string]any{"uuid": "fixture-adapter", "name": "fixture", "status": "DRAFT", "created_at": "2026-10-04", "updated_at": "2026-10-04"}}, "pagination": map[string]any{"total_items": 1}})
			return
		}
		if saved == nil {
			w.WriteHeader(404)
			send(map[string]any{"message": "gone"})
			return
		}
		remote := map[string]any{}
		for k, v := range saved {
			remote[k] = v
		}
		remote["uuid"] = "fixture-adapter"
		remote["status"] = "DRAFT"
		remote["created_at"] = "2026-10-04"
		remote["updated_at"] = "2026-10-04"
		values := []any{}
		for _, item := range saved["variables"].([]any) {
			v := item.(map[string]any)
			copy := map[string]any{}
			for k, x := range v {
				copy[k] = x
			}
			if v["type"] == "SECRET" {
				copy["value"] = nil
				copy["is_redacted"] = true
			}
			values = append(values, copy)
		}
		remote["variables"] = values
		send(remote)
	}))
	defer server.Close()
	provider := fmt.Sprintf(`provider "prisma-airs" {
 client_id="fixture"
 client_secret="fixture"
 tsg_id="123"
 token_endpoint=%q
 red_team {mgmt_endpoint=%q}
}
`, server.URL+"/token", server.URL)
	config := func(secret, plain string) string {
		return provider + fmt.Sprintf(`resource "prisma-airs_red_team_adapter" "test" {
 name="fixture"
 script="print('fixture')"
 variables={credential={type="SECRET",value=%s},endpoint={type="VAR",value=%q}}
}
data "prisma-airs_red_team_adapters" "all" {depends_on=[prisma-airs_red_team_adapter.test]}
data "prisma-airs_red_team_adapter" "existing" {id=data.prisma-airs_red_team_adapters.all.ids_by_name["fixture"]}
`, secret, plain)
	}
	const addr = "prisma-airs_red_team_adapter.test"
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config(`"original-secret"`, "first")},
		{ResourceName: addr, ImportState: true, ImportStateId: "fixture-adapter", ImportStateVerify: true, ImportStateVerifyIgnore: []string{"variables"}},
		{Config: config("null", "second")},
		{Config: config("null", "second"), PlanOnly: true},
		{Config: config(`"********"`, "third"), ExpectError: regexp.MustCompile("Masked adapter secret")},
	}})
	if writes != 2 || storedSecret != "original-secret" {
		t.Fatalf("unexpected writes or retained secret: %d", writes)
	}
}
