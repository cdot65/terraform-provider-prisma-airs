package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const skillRuleID = "12345678-1234-1234-1234-123456789abc"
const skillOverrideID = "87654321-4321-4321-4321-cba987654321"
const skillScanID = "35f902a2-8b23-4d44-8e6b-466c4d927534"

type skillFixture struct {
	sync.Mutex
	instance, override map[string]any
	ruleState          string
	ruleInstanceUUID   string
	writes             []map[string]any
	deleted            int
	failRead           bool
	normalizeInstance  bool
	failUpdate         bool
	foreignInstance    bool
}

func newSkillFixture(t *testing.T) (*skillFixture, *httptest.Server) {
	t.Helper()
	f := &skillFixture{ruleState: "ALLOWING", ruleInstanceUUID: skillOverrideID}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.Lock()
		defer f.Unlock()
		w.Header().Set("Content-Type", "application/json")
		send := func(v any) {
			if e := json.NewEncoder(w).Encode(v); e != nil {
				t.Error(e)
			}
		}
		if r.URL.Path == "/token" {
			if r.Header.Get("x-tsg-id") != "" {
				t.Error("tenant leaked to OAuth")
			}
			send(map[string]any{"access_token": "test-token", "expires_in": 3600})
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("x-tsg-id") != "123" {
			t.Error("missing OAuth/tenant headers")
		}
		if r.Method == "GET" && f.failRead {
			w.WriteHeader(403)
			send(map[string]string{"message": "private-secret"})
			return
		}
		rule := map[string]any{"uuid": skillRuleID, "name": "Secrets", "default_state": "BLOCKING", "vulnerability_types": []string{"SECRET_EXPOSURE"}, "description": "", "remediation": ""}
		rules := []any{rule}
		instances := []any{map[string]any{"uuid": f.ruleInstanceUUID, "rule_uuid": skillRuleID, "security_group_uuid": skillRuleID, "tsg_id": "123", "state": f.ruleState, "rule": rule, "created_at": "", "updated_at": ""}}
		if r.URL.Query().Get("skip") != "" && r.URL.Query().Get("skip") != "0" {
			rules = []any{}
			instances = []any{}
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/mgmt/v1/instances"):
			if r.Method == "GET" {
				if f.instance == nil {
					w.WriteHeader(404)
					send(map[string]string{"message": "absent"})
					return
				}
				if f.normalizeInstance {
					copy := map[string]any{"tenant_id": f.instance["tenant_id"], "tsg_id": f.instance["tsg_id"], "created_by": "service-audit", "support_account_id": "normalized-support"}
					send(copy)
				} else if f.foreignInstance {
					send(map[string]any{"tenant_id": "tenant/a", "tsg_id": "foreign", "support_account_id": "support"})
				} else {
					send(f.instance)
				}
				return
			}
			if r.Method == "DELETE" {
				f.instance = nil
				f.deleted++
				send(map[string]any{"is_success": true, "tenant_id": "tenant/a", "tsg_id": "123"})
				return
			}
			if r.Method == "PUT" && f.failUpdate {
				w.WriteHeader(400)
				send(map[string]string{"message": "private-secret"})
				return
			}
			var body map[string]any
			if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
				t.Error(e)
			}
			f.writes = append(f.writes, body)
			f.instance = map[string]any{"tenant_id": body["tenant_id"], "tsg_id": body["tsg_id"], "created_by": body["created_by"], "support_account_id": body["support_account_id"], "support_account_name": body["support_account_name"]}
			send(map[string]any{"is_success": true, "tenant_id": body["tenant_id"], "tsg_id": body["tsg_id"]})
		case r.URL.Path == "/mgmt/v1/rules":
			send(map[string]any{"rules": rules, "pagination": map[string]int{"total_items": len(rules)}})
		case r.URL.Path == "/mgmt/v1/rule-instances":
			if r.Method == "PUT" {
				var body struct {
					Rules map[string]struct {
						State string `json:"state"`
					} `json:"rule_configurations"`
				}
				if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
					t.Error(e)
				}
				if len(body.Rules) != 1 {
					t.Error("unowned rules modified")
				}
				f.ruleState = body.Rules[skillRuleID].State
				f.writes = append(f.writes, map[string]any{"state": f.ruleState})
				if f.ruleInstanceUUID == skillOverrideID {
					f.ruleInstanceUUID = skillScanID
				} else {
					f.ruleInstanceUUID = skillOverrideID
				}
				instances[0].(map[string]any)["uuid"] = f.ruleInstanceUUID
				instances[0].(map[string]any)["state"] = f.ruleState
			}
			send(map[string]any{"rule_instances": instances, "pagination": map[string]int{"total_items": len(instances)}})
		case strings.HasPrefix(r.URL.Path, "/mgmt/v1/skill-overrides"):
			if r.Method == "POST" {
				var body map[string]any
				if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
					t.Error(e)
				}
				f.writes = append(f.writes, body)
				f.override = body
				if body["reason"] == "" {
					f.override["reason"] = nil
				}
				f.override["uuid"] = skillOverrideID
				f.override["tsg_id"] = "123"
				f.override["created_at"] = "2026-10-03T00:00:00Z"
				f.override["updated_at"] = "2026-10-03T00:00:00Z"
				f.override["created_by"] = "user"
				send(f.override)
				return
			}
			if r.Method == "DELETE" {
				f.override = nil
				f.deleted++
				w.WriteHeader(204)
				return
			}
			rows := []any{}
			if f.override != nil && (r.URL.Query().Get("skip") == "" || r.URL.Query().Get("skip") == "0") {
				rows = append(rows, f.override)
			}
			send(map[string]any{"skill_overrides": rows, "pagination": map[string]int{"total_items": len(rows)}})
		case strings.HasSuffix(r.URL.Path, "/vulnerabilities"):
			send(map[string]any{"vulnerabilities": []any{}, "pagination": map[string]int{"total_items": 0}})
		case strings.Contains(r.URL.Path, "/attack-chains"):
			send(map[string]any{"attack_chains": []any{}, "pagination": map[string]int{"total_items": 0}})
		case r.URL.Path == "/data/v1/stats/scans":
			send(map[string]any{"unique_skills_scanned": map[string]any{"count": 1, "percent_change": nil}, "total_vulnerabilities_found": map[string]any{"count": 0, "percent_change": nil}})
		case r.URL.Path == "/data/v1/stats/rules":
			send(map[string]any{"total_enabled_rules": nil, "total_rule_violations": map[string]any{"count": 0, "percent_change": nil}, "most_violated_rules": nil})
		case strings.HasPrefix(r.URL.Path, "/data/v1/scans"):
			scan := map[string]any{"uuid": skillScanID, "tsg_id": "123", "name": "fixture", "status": "COMPLETED", "created_at": "2026-10-03T00:00:00Z", "updated_at": "2026-10-03T00:00:00Z", "eval_outcome": "ALLOWED"}
			if r.URL.Path == "/data/v1/scans" {
				send(map[string]any{"scans": []any{scan}, "pagination": map[string]int{"total_items": 1}})
			} else {
				send(scan)
			}
		default:
			t.Errorf("unexpected endpoint %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(server.Close)
	return f, server
}
func skillProviderConfig(server *httptest.Server) string {
	return fmt.Sprintf(`provider "prisma-airs" {
 client_id="client"
 client_secret="secret"
 tsg_id="123"
 token_endpoint=%q
 supply_chain {
  skill_scanning_data_endpoint=%q
  skill_scanning_mgmt_endpoint=%q
 }
}
`, server.URL+"/token", server.URL+"/data", server.URL+"/mgmt")
}
func TestSkillScanningTerraformInstanceCRUD(t *testing.T) {
	f, server := newSkillFixture(t)
	base := skillProviderConfig(server)
	addr := "prisma-airs_supply_chain_skill_scanning_instance.test"
	config := func(name, extras string) string {
		return base + fmt.Sprintf(`resource "prisma-airs_supply_chain_skill_scanning_instance" "test" {
 tenant_id="tenant/a"
 support_account_id="support"
 created_by="user@example.com"
 registration_details={
  region="us"
  license_name="desired-license"
  entitlements=[]
  tsg_instances=[]
 }
 support_account_name=%q
 %s
}
`, name, extras)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("First", `auth_code="private-code"
auth_code_version=1
iam_controlled=false`), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "id", "tenant/a"), resource.TestCheckResourceAttr(addr, "iam_controlled", "false"), resource.TestCheckNoResourceAttr(addr, "auth_code"))},
		{Config: config("Updated", "auth_code_version=2"), Check: resource.TestCheckResourceAttr(addr, "support_account_name", "Updated")},
		{Config: config("Metadata update", "auth_code_version=2"), Check: resource.TestCheckResourceAttr(addr, "support_account_name", "Metadata update")},
		{ResourceName: addr, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"auth_code", "auth_code_version", "iam_controlled", "registration_details"}},
	}, CheckDestroy: func(*terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if f.instance != nil || f.deleted != 1 {
			return fmt.Errorf("instance was not destroyed")
		}
		_, iamPresent := f.writes[1]["iam_controlled"]
		_, codePresent := f.writes[1]["auth_code"]
		if !iamPresent || !codePresent {
			return fmt.Errorf("clear keys were omitted instead of explicit null")
		}
		if len(f.writes) != 3 || f.writes[0]["iam_controlled"] != false || f.writes[1]["iam_controlled"] != nil || f.writes[1]["auth_code"] != nil {
			return fmt.Errorf("false/null intent was lost")
		}
		for _, body := range f.writes {
			if body["region"] != "us" || body["license_name"] != "desired-license" {
				return fmt.Errorf("full desired registration was dropped from PUT")
			}
		}
		if _, present := f.writes[2]["auth_code"]; present {
			return fmt.Errorf("authorization code was resent without a version change")
		}
		return nil
	}})
}
func TestSkillScanningTerraformRuleUpdatesAndRestoration(t *testing.T) {
	f, server := newSkillFixture(t)
	base := skillProviderConfig(server)
	addr := "prisma-airs_supply_chain_skill_scanning_rule.test"
	config := func(state string) string {
		return base + fmt.Sprintf(`resource "prisma-airs_supply_chain_skill_scanning_rule" "test" {
 rule_uuid=%q
 state=%q
}
`, skillRuleID, state)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("BLOCKING"), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "original_state", "ALLOWING"), resource.TestCheckResourceAttr(addr, "id", skillRuleID), resource.TestCheckResourceAttr(addr, "rule_instance_uuid", skillScanID))},
		{Config: config("DISABLED"), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "original_state", "ALLOWING"), resource.TestCheckResourceAttr(addr, "id", skillRuleID), resource.TestCheckResourceAttr(addr, "rule_instance_uuid", skillOverrideID))},
		// Import deliberately captures the current state, changing the restoration
		// baseline; do not replace the existing managed state in this test.
		{ResourceName: addr, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"original_state"}},
	}, CheckDestroy: func(*terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if f.ruleState != "ALLOWING" || len(f.writes) != 3 {
			return fmt.Errorf("captured policy was not restored")
		}
		return nil
	}})
}
func TestSkillScanningTerraformOverrideReplacementImportAndDelete(t *testing.T) {
	f, server := newSkillFixture(t)
	base := skillProviderConfig(server)
	addr := "prisma-airs_supply_chain_skill_scanning_override.test"
	config := func(reason string) string {
		return base + fmt.Sprintf(`resource "prisma-airs_supply_chain_skill_scanning_override" "test" {
 skill_name="fixture"
 fingerprint=%q
 trusted_by="user@example.com"
 reason=%q
}
`, strings.Repeat("a", 64), reason)
	}
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("first"), Check: resource.TestCheckResourceAttr(addr, "decision", "ALLOW")},
		{ResourceName: addr, ImportState: true, ImportStateVerify: true},
		{Config: config(""), Check: resource.TestCheckResourceAttr(addr, "reason", "")},
	}, CheckDestroy: func(*terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if f.override != nil || f.deleted != 2 || len(f.writes) != 2 {
			return fmt.Errorf("override replacement/delete lifecycle is incorrect")
		}
		return nil
	}})
}
func TestSkillScanningTerraformDataSources(t *testing.T) {
	_, server := newSkillFixture(t)
	base := skillProviderConfig(server)
	config := base + fmt.Sprintf(`
data "prisma-airs_supply_chain_skill_scanning_rules" "test" {}
data "prisma-airs_supply_chain_skill_scanning_rule_instances" "test" {}
data "prisma-airs_supply_chain_skill_scanning_overrides" "test" { q="fixture" }
data "prisma-airs_supply_chain_skill_scanning_scan" "test" { scan_uuid=%q }
data "prisma-airs_supply_chain_skill_scanning_scan" "lookup" { fingerprint=%q }
data "prisma-airs_supply_chain_skill_scanning_scans" "test" { statuses=["COMPLETED","FAILED"] }
data "prisma-airs_supply_chain_skill_scanning_vulnerabilities" "test" {
 scan_uuid=%q
 in_chain=false
}
data "prisma-airs_supply_chain_skill_scanning_attack_chains" "test" { scan_uuid=%q }
data "prisma-airs_supply_chain_skill_scanning_statistics" "test" { time_period="7_DAYS" }
output "rule_count" { value=length(data.prisma-airs_supply_chain_skill_scanning_rules.test.result.rules) }
output "unavailable_rule_count" { value=data.prisma-airs_supply_chain_skill_scanning_statistics.test.result.rules.total_enabled_rules == null }
output "zero_findings" { value=data.prisma-airs_supply_chain_skill_scanning_statistics.test.result.scans.total_vulnerabilities_found.count }
output "scan_status" {
 value=data.prisma-airs_supply_chain_skill_scanning_scan.test.result.status
 sensitive=true
}
`, skillScanID, strings.Repeat("a", 64), skillScanID, skillScanID)
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{{Config: config, Check: resource.ComposeTestCheckFunc(resource.TestCheckOutput("rule_count", "1"), resource.TestCheckOutput("scan_status", "COMPLETED"), resource.TestCheckOutput("unavailable_rule_count", "true"), resource.TestCheckOutput("zero_findings", "0"))}}})
}
func TestSkillScanningTerraformValidatesSelectors(t *testing.T) {
	_, server := newSkillFixture(t)
	config := skillProviderConfig(server) + `data "prisma-airs_supply_chain_skill_scanning_scan" "test" {}`
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile(`(?i)one.*only one`)}}})
}

func TestSkillScanningTerraformRejectsInvalidFilters(t *testing.T) {
	for _, body := range []string{
		`data "prisma-airs_supply_chain_skill_scanning_scans" "test" {
 start_time="2026-10-03T00:00:00Z"
 end_time="2026-10-02T00:00:00Z"
}`,
		`data "prisma-airs_supply_chain_skill_scanning_attack_chains" "test" {
 scan_uuid="35f902a2-8b23-4d44-8e6b-466c4d927534"
 chain_uuid="12345678-1234-1234-1234-123456789abc"
 limit=10
}`,
		`data "prisma-airs_supply_chain_skill_scanning_vulnerabilities" "test" {
 scan_uuid="35f902a2-8b23-4d44-8e6b-466c4d927534"
 type="TYPO"
}`,
		`data "prisma-airs_supply_chain_skill_scanning_scans" "test" { fingerprint="invalid" }`,
		`data "prisma-airs_supply_chain_skill_scanning_overrides" "test" { fingerprint="invalid" }`,
		`data "prisma-airs_supply_chain_skill_scanning_overrides" "test" { trusted_by="` + strings.Repeat("x", 257) + `" }`,
		`data "prisma-airs_supply_chain_skill_scanning_vulnerabilities" "test" {
 scan_uuid="35f902a2-8b23-4d44-8e6b-466c4d927534"
 limit=101
}`,
		`resource "prisma-airs_supply_chain_skill_scanning_override" "test" {
 skill_name="invalid"
 fingerprint="AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
 trusted_by="user@example.com"
}`,
	} {
		t.Run(body, func(t *testing.T) {
			_, server := newSkillFixture(t)
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{{Config: skillProviderConfig(server) + body, ExpectError: regexp.MustCompile(`(?i)invalid|must|conflict`)}}})
		})
	}
}

func TestSkillScanningTerraformNormalizedInstanceFailedUpdateAndReplan(t *testing.T) {
	f, server := newSkillFixture(t)
	f.normalizeInstance = true
	base := skillProviderConfig(server)
	addr := "prisma-airs_supply_chain_skill_scanning_instance.test"
	config := func(name string) string {
		return base + fmt.Sprintf(`resource "prisma-airs_supply_chain_skill_scanning_instance" "test" {
 tenant_id="tenant/a"
 support_account_id="desired-support"
 created_by="desired-user"
 support_account_name=%q
 registration_details={region="us",tsg_instances=[],entitlements=[]}
}`, name)
	}
	setFailure := func(v bool) func() { return func() { f.Lock(); defer f.Unlock(); f.failUpdate = v } }
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("First"), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "created_by", "desired-user"), resource.TestCheckResourceAttr(addr, "support_account_id", "desired-support"))},
		{PreConfig: setFailure(true), Config: config("Updated"), ExpectError: regexp.MustCompile(`HTTP 400`)},
		{PreConfig: setFailure(false), Config: config("Updated"), PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config("Updated"), Check: resource.TestCheckResourceAttr(addr, "support_account_name", "Updated")},
		{Config: config("Updated"), PlanOnly: true},
	}})
}
func TestSkillScanningTerraformInstanceDataRejectsForeignTenant(t *testing.T) {
	f, server := newSkillFixture(t)
	f.instance = map[string]any{"tenant_id": "tenant/a"}
	f.foreignInstance = true
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{{Config: skillProviderConfig(server) + `data "prisma-airs_supply_chain_skill_scanning_instance" "test" {tenant_id="tenant/a"}`, ExpectError: regexp.MustCompile(`instance response does not match`)}}})
}

func TestSkillScanningTerraformInstanceImportRequiresFirstPUT(t *testing.T) {
	f, server := newSkillFixture(t)
	f.instance = map[string]any{"tenant_id": "tenant/a", "tsg_id": "123", "support_account_id": "support", "created_by": "user"}
	addr := "prisma-airs_supply_chain_skill_scanning_instance.test"
	config := skillProviderConfig(server) + `resource "prisma-airs_supply_chain_skill_scanning_instance" "test" {
 tenant_id="tenant/a"
 support_account_id="support"
 created_by="user"
 registration_details={region="us",entitlements=[],tsg_instances=[]}
}`
	resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, ResourceName: addr, ImportStateId: "tenant/a", ImportState: true, ImportStatePersist: true},
		{Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config},
		{Config: config, PlanOnly: true},
	}, CheckDestroy: func(*terraform.State) error {
		f.Lock()
		defer f.Unlock()
		if len(f.writes) != 1 || f.writes[0]["region"] != "us" {
			return fmt.Errorf("import did not reconcile complete registration on first apply")
		}
		return nil
	}})
}
