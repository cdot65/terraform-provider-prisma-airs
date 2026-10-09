package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// This mock replays the sanitized 2026-10-09 response, not a live API certification.
type securityProfileFixture struct {
	failCreationMetadata bool
	readCalls            int
	failReads            bool
	sync.Mutex
	profiles       []airsruntime.SecurityProfile
	fixture        airsruntime.SecurityProfile
	writes         int
	expectedPolicy *airsruntime.ProfilePolicy
	legacy         bool
}

func newSecurityProfileFixture(t *testing.T) (*securityProfileFixture, *httptest.Server) {
	t.Helper()
	f := &securityProfileFixture{}
	raw, err := os.ReadFile("../products/runtime/testdata/directional-security-profile.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &f.fixture); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		f.Lock()
		defer f.Unlock()
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
		if q.Header.Get("Authorization") != "Bearer fixture" {
			t.Error("OAuth token missing")
		}
		if q.Method == http.MethodGet && q.URL.Path == "/v1/mgmt/profiles/tsg/100" {
			f.readCalls++
			if f.failReads || (f.failCreationMetadata && f.readCalls%2 == 0) {
				w.WriteHeader(400)
				send(map[string]any{"message": "fixture read failure"})
				return
			}
			rows := make([]airsruntime.SecurityProfile, len(f.profiles))
			copy(rows, f.profiles)
			for i := range rows {
				rows[i].DLPTenantID = ""
				rows[i].SetFieldPresence("dlp_tenant_id", airsruntime.JSONOmitted)
			}
			send(map[string]any{"ai_profiles": rows})
			return
		}
		if q.Method == http.MethodPost || q.Method == http.MethodPut {
			var body airsruntime.CreateProfileRequest
			if err := json.NewDecoder(q.Body).Decode(&body); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			if q.Method == http.MethodPost && q.URL.Path != "/v1/mgmt/profile" {
				t.Errorf("unexpected create path %s", q.URL.Path)
			}
			if q.Method == http.MethodPost && !f.legacy {
				assertLifecyclePolicy(t, configuredFixturePolicy(t, f.fixture.Policy), body.Policy)
			}
			if q.Method == http.MethodPut {
				if len(f.profiles) == 0 || q.URL.Path != "/v1/mgmt/profile/uuid/"+f.profiles[len(f.profiles)-1].ProfileID {
					t.Error("update did not use latest revision UUID")
				}
				if f.expectedPolicy != nil {
					assertLifecyclePolicy(t, f.expectedPolicy, body.Policy)
				}
				if !reflect.DeepEqual(f.profiles[len(f.profiles)-1].Extensions, body.Extensions) {
					t.Error("top-level profile extension lost")
				}
			}
			p := f.fixture
			p.ProfileName = body.ProfileName
			f.writes++
			p.ProfileID = fmt.Sprintf("550e8400-e29b-41d4-a716-%012d", f.writes)
			p.Revision = int32(f.writes)
			if q.Method == http.MethodPut || f.legacy {
				p.Policy = body.Policy
				p.Extensions = body.Extensions
			}
			f.profiles = append(f.profiles, p)
			send(p)
			return
		}
		if q.Method == http.MethodDelete {
			for i, p := range f.profiles {
				if strings.Contains(q.URL.Path, p.ProfileID) {
					f.profiles = append(f.profiles[:i], f.profiles[i+1:]...)
					send("deleted")
					return
				}
			}
			w.WriteHeader(404)
			send(map[string]any{"message": "absent"})
			return
		}
		t.Errorf("unexpected API request %s %s", q.Method, q.URL.Path)
		w.WriteHeader(400)
	}))
	t.Cleanup(server.Close)
	return f, server
}
func assertLifecyclePolicy(t *testing.T, want, got *airsruntime.ProfilePolicy) {
	t.Helper()
	tree := func(value any) any {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		if err = json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	}
	if !reflect.DeepEqual(tree(want), tree(got)) {
		a, _ := json.Marshal(want)
		b, _ := json.Marshal(got)
		t.Errorf("write changed unrelated policy\nwant %s\ngot %s", a, b)
	}
}
func securityProfileProviderConfig(server *httptest.Server) string {
	return fmt.Sprintf(`provider "prisma-airs" {
 client_id="fixture-client"
 client_secret="fixture-secret"
 tsg_id="100"
 token_endpoint=%q
 runtime { mgmt_endpoint=%q }
}
`, server.URL+"/token", server.URL)
}
func TestSecurityProfileDirectionalTerraformLifecycle(t *testing.T) {
	f, server := newSecurityProfileFixture(t)
	if err := f.fixture.SetExtension("future-profile", json.RawMessage(`{"enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	// Future direction and detector fields must survive separate Terraform processes.
	if err := f.fixture.Policy.AiSecurityProfiles[0].ContentTypeConfigurations.SetExtension("future-direction", json.RawMessage(`{"enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	if err := f.fixture.Policy.AiSecurityProfiles[0].ContentTypeConfigurations.Response.ModelProtection[0].SetExtension("future-setting", json.RawMessage(`{"limit":9007199254740993}`)); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("testdata/directional-security-profile.tf")
	if err != nil {
		t.Fatal(err)
	}
	config := securityProfileProviderConfig(server) + string(raw)
	edited := config
	// Only response toxicity changes; no other direction's override changes.
	start := strings.Index(edited, "      response {")
	end := strings.Index(edited[start:], "      tool_call {") + start
	edited = edited[:start] + strings.Replace(edited[start:end], `high     = "medium"`, `high     = "critical"`, 1) + edited[end:]
	if edited == config {
		t.Fatal("fixture response severity edit did not match HCL")
	}
	address := "prisma-airs_runtime_security_profile.directional"
	prepareExpected := func() {
		f.Lock()
		defer f.Unlock()
		raw, err := json.Marshal(f.fixture.Policy)
		if err != nil {
			t.Fatal(err)
		}
		var p airsruntime.ProfilePolicy
		if err = json.Unmarshal(raw, &p); err != nil {
			t.Fatal(err)
		}
		p.AiSecurityProfiles[0].ContentTypeConfigurations.Response.ModelProtection[0].SeverityByConfidence.High = "critical"
		f.expectedPolicy = &p
	}
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(address, "ai_security_profile.0.content_type_mode", "per_content_type"),
		resource.TestCheckResourceAttr(address, "ai_security_profile.0.enable_full_conversation_inspection", "false"),
		resource.TestCheckResourceAttr(address, "ai_security_profile.0.content_type_configurations.prompt.data_protection.data_leak_detection.member.0.id", ""),
		resource.TestCheckResourceAttr(address, "ai_security_profile.0.content_type_configurations.response.model_protection.0.severity_by_confidence.high", "medium"),
	)
	global := strings.Replace(edited, "max_inline_latency    = 5", "max_inline_latency    = 9", 1)
	removed := strings.Replace(global, `        model_protection {
          action   = "block"
          name     = "prompt-injection"
          severity = "medium"
        }
`, "", 1)
	if removed == global || global == edited {
		t.Fatal("fixture global/removal edit failed")
	}
	expectedFromLatest := func(edit func(*airsruntime.ProfilePolicy)) func() {
		return func() {
			f.Lock()
			defer f.Unlock()
			raw, err := json.Marshal(f.profiles[len(f.profiles)-1].Policy)
			if err != nil {
				t.Fatal(err)
			}
			var p airsruntime.ProfilePolicy
			if err = json.Unmarshal(raw, &p); err != nil {
				t.Fatal(err)
			}
			edit(&p)
			f.expectedPolicy = &p
		}
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: checks},
		{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{ResourceName: address, ImportState: true, ImportStateId: "Directionality Test", ImportStateVerify: true, ImportStateVerifyIgnore: []string{"dlp_tenant_id"}},
		{Config: edited, PreConfig: prepareExpected, Check: resource.TestCheckResourceAttr(address, "revision", "2")},
		{Config: edited, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{Config: global, PreConfig: expectedFromLatest(func(p *airsruntime.ProfilePolicy) {
			p.AiSecurityProfiles[0].ModelConfiguration.Latency.MaxInlineLatency = 9
		}), Check: resource.TestCheckResourceAttr(address, "revision", "3")},
		{Config: removed, PreConfig: expectedFromLatest(func(p *airsruntime.ProfilePolicy) {
			p.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection = p.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection[1:]
		}), Check: resource.TestCheckResourceAttr(address, "ai_security_profile.0.content_type_configurations.prompt.model_protection.#", "1")},
		{Config: removed, PlanOnly: true, PreConfig: func() {
			f.Lock()
			f.profiles[len(f.profiles)-1].Policy.AiSecurityProfiles[0].ContentTypeConfigurations.Response.ModelProtection[0].SeverityByConfidence.High = "low"
			f.Unlock()
		}, ExpectNonEmptyPlan: true},
		{Config: removed, PreConfig: expectedFromLatest(func(p *airsruntime.ProfilePolicy) {
			p.AiSecurityProfiles[0].ContentTypeConfigurations.Response.ModelProtection[0].SeverityByConfidence.High = "critical"
		}), Check: resource.TestCheckResourceAttr(address, "revision", "5")},
		{Config: removed, PlanOnly: true, PreConfig: func() { f.Lock(); f.failReads = true; f.Unlock() }, ExpectError: regexp.MustCompile("Failed to read security profile")},
		{Config: removed, PreConfig: func() { f.Lock(); f.failReads = false; f.Unlock() }, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, Check: resource.TestCheckResourceAttr(address, "revision", "5")},
		{Config: removed, PlanOnly: true, PreConfig: func() { f.Lock(); f.failCreationMetadata = true; f.readCalls = 0; f.Unlock() }, ExpectError: regexp.MustCompile("Failed to read profile creation metadata")},
		{Config: removed, PreConfig: func() { f.Lock(); f.failCreationMetadata = false; f.Unlock() }, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, Check: resource.TestCheckResourceAttr(address, "revision", "5")},
	}})
}
func TestSecurityProfileLegacyTerraformLifecycle(t *testing.T) {
	f, server := newSecurityProfileFixture(t)
	f.legacy = true
	config := securityProfileProviderConfig(server) + `resource "prisma-airs_runtime_security_profile" "legacy" {
 profile_name="legacy"
 ai_security_profile {
  model_protection {
   name="prompt-injection"
   action="block"
  }
 }
}`
	address := "prisma-airs_runtime_security_profile.legacy"
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: func(_ *terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if len(f.profiles) != 0 {
			return fmt.Errorf("history remains")
		}
		return nil
	}, Steps: []resource.TestStep{
		{Config: config, Check: resource.TestCheckResourceAttr(address, "ai_security_profile.0.model_protection.0.action", "block")},
		{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{ResourceName: address, ImportState: true, ImportStateId: "legacy", ImportStateVerify: true, ImportStateVerifyIgnore: []string{"dlp_tenant_id"}},
		{Config: strings.Replace(config, `action="block"`, `action="allow"`, 1), Check: resource.TestCheckResourceAttr(address, "revision", "2")},
	}})
}

func TestSecurityProfileDirectionalImportConverges(t *testing.T) {
	f, server := newSecurityProfileFixture(t)
	f.profiles = []airsruntime.SecurityProfile{f.fixture}
	f.writes = 1
	if err := f.fixture.Policy.AiSecurityProfiles[0].ContentTypeConfigurations.SetExtension("future-direction", json.RawMessage(`{"enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("testdata/directional-security-profile.tf")
	if err != nil {
		t.Fatal(err)
	}
	config := securityProfileProviderConfig(server) + string(raw)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{ResourceName: "prisma-airs_runtime_security_profile.directional", ImportState: true, ImportStateId: "Directionality Test", ImportStatePersist: true},
		{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, Check: resource.TestCheckResourceAttr("prisma-airs_runtime_security_profile.directional", "ai_security_profile.0.content_type_configurations.response.model_protection.0.severity_by_confidence.high", "medium")},
	}})
}

// Derive the create expectation from the fixture and the authored HCL. Empty URL
// objects, empty agent arrays, prompt database null and future fields are added
// by the mock service; every explicitly configured setting must be sent on POST.
func configuredFixturePolicy(t *testing.T, fixture *airsruntime.ProfilePolicy) *airsruntime.ProfilePolicy {
	t.Helper()
	raw, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var policy airsruntime.ProfilePolicy
	if err = json.Unmarshal(raw, &policy); err != nil {
		t.Fatal(err)
	}
	policy.SetFieldPresence("dlp-data-profiles", airsruntime.JSONOmitted)
	directions := policy.AiSecurityProfiles[0].ContentTypeConfigurations
	directions.Extensions = nil
	for _, p := range []*airsruntime.ProtectionConfiguration{directions.Prompt, directions.Response, directions.ToolCall, directions.ToolResponse} {
		p.SetFieldPresence("agent-protection", airsruntime.JSONOmitted)
		if p.DataProtection.DatabaseSecurity == nil {
			p.DataProtection.SetFieldPresence("database-security", airsruntime.JSONOmitted)
		}
		p.AppProtection.AlertURLCategory = nil
		p.AppProtection.AllowURLCategory = nil
		p.AppProtection.BlockURLCategory = nil
		for i := range p.ModelProtection {
			p.ModelProtection[i].Extensions = nil
		}
	}
	return &policy
}
