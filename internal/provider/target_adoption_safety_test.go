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

func TestTargetAdoptionBlocksIncompleteWritesBeforeAPI(t *testing.T) {
	var lock sync.Mutex
	writes := 0
	deleted := false
	description := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		send := func(value any) {
			if err := json.NewEncoder(w).Encode(value); err != nil {
				t.Error(err)
			}
		}
		if q.URL.Path == "/token" {
			send(map[string]any{"access_token": "fixture", "expires_in": 3600})
			return
		}
		if q.URL.Path != "/v1/target/fixture-target" {
			t.Errorf("unexpected request %s %s", q.Method, q.URL.Path)
			w.WriteHeader(404)
			return
		}
		if q.Method == "DELETE" {
			deleted = true
			send(map[string]any{})
			return
		}
		if deleted {
			w.WriteHeader(404)
			send(map[string]any{"message": "gone"})
			return
		}
		if q.Method == "PUT" {
			writes++
			var body map[string]any
			if err := json.NewDecoder(q.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			description = body["description"].(string)
			headers := body["connection_params"].(map[string]any)["request_headers"].(map[string]any)
			if headers["Authorization"] != "original" || q.URL.Query().Get("validate") != "false" {
				t.Error("incomplete credential or inference-enabled write reached API")
			}
		}
		send(map[string]any{"uuid": "fixture-target", "name": "fixture", "description": description,
			"target_type": "APPLICATION", "connection_type": "CUSTOM", "response_mode": "REST", "api_endpoint_type": "PUBLIC", "status": "DRAFT", "created_at": "2026-10-04", "updated_at": "2026-10-04",
			"connection_params": map[string]any{"api_endpoint": "https://example.com", "request_json": map[string]any{"prompt": "{INPUT}"}, "response_json": map[string]any{"output": "{RESPONSE}"}, "response_key": "output", "request_headers": map[string]any{"Authorization": "********"}}})
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
	config := func(family, headers, description string) string {
		return provider + fmt.Sprintf(`import {
 to=prisma-airs_red_team_target.test
 id="rest/fixture-target"
}
resource "prisma-airs_red_team_target" "test" {
 name="fixture"
 description=%q
 %s {
 api_endpoint="https://example.com"
 request_body={prompt="{INPUT}"}
 response_body={output="{RESPONSE}"}
 response_key="output"
 request_headers=%s
 }
}
`, description, family, headers)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("rest", "null", "")},
		{Config: config("rest", "null", ""), PlanOnly: true},
		{Config: config("rest", "{}", "edited"), ExpectError: regexp.MustCompile("Unavailable imported target input")},
		{Config: config("custom", "{}", "edited"), ExpectError: regexp.MustCompile("Unavailable imported target input")},
		{Config: config("custom", `{Authorization="original"}`, "edited")},
		{Config: config("custom", `{Authorization="original"}`, "edited"), PlanOnly: true},
	}})
	if writes != 1 {
		t.Fatalf("expected only the complete update to reach API; got %d writes", writes)
	}
}
