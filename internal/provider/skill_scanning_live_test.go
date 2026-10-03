package provider_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	ag "github.com/cdot65/prisma-airs-go/aisec/agentguard"
	s "github.com/cdot65/prisma-airs-go/aisec/agentguard/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func skillLiveClient(t *testing.T) *ag.Client {
	t.Helper()
	if os.Getenv("PANW_SKILL_SCANNING_DATA_ENDPOINT") == "" || os.Getenv("PANW_SKILL_SCANNING_MGMT_ENDPOINT") == "" {
		t.Skip("live Skill Scanning tests require both PANW_SKILL_SCANNING endpoint variables")
	}
	c, e := ag.NewClient(ag.Opts{ClientID: os.Getenv("PANW_MGMT_CLIENT_ID"), ClientSecret: os.Getenv("PANW_MGMT_CLIENT_SECRET"), TsgID: os.Getenv("PANW_MGMT_TSG_ID"), TokenEndpoint: os.Getenv("PANW_MGMT_TOKEN_ENDPOINT"), DataEndpoint: os.Getenv("PANW_SKILL_SCANNING_DATA_ENDPOINT"), MgmtEndpoint: os.Getenv("PANW_SKILL_SCANNING_MGMT_ENDPOINT")})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func skillLiveConfig() string {
	return fmt.Sprintf(`provider "prisma-airs" {
 supply_chain {
  skill_scanning_data_endpoint=%q
  skill_scanning_mgmt_endpoint=%q
 }
}
`, os.Getenv("PANW_SKILL_SCANNING_DATA_ENDPOINT"), os.Getenv("PANW_SKILL_SCANNING_MGMT_ENDPOINT"))
}
func skillLiveOverrides(ctx context.Context, c *ag.Client) ([]s.SkillOverrideResponse, error) {
	rows := []s.SkillOverrideResponse{}
	seen := map[string]bool{}
	for page := 0; page < 100; page++ {
		p, e := c.SkillOverrides.List(ctx, ag.SkillOverrideListOpts{ListOpts: ag.ListOpts{Limit: 100, Skip: len(rows)}})
		if e != nil {
			return nil, e
		}
		if len(p.SkillOverrides) == 0 {
			return rows, nil
		}
		for _, item := range p.SkillOverrides {
			if item.UUID == "" || seen[item.UUID] {
				return nil, fmt.Errorf("duplicate/missing override UUID")
			}
			seen[item.UUID] = true
		}
		rows = append(rows, p.SkillOverrides...)
	}
	return nil, fmt.Errorf("override inventory exceeded budget")
}
func TestAccSkillScanningTrustedOverride(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("live test requires TF_ACC=1")
	}
	testAccPreCheck(t)
	c := skillLiveClient(t)
	raw := make([]byte, 32)
	if _, e := rand.Read(raw); e != nil {
		t.Fatal(e)
	}
	fingerprint := hex.EncodeToString(raw)
	name := fmt.Sprintf("terraform-skill-scanning-e2e-%d", time.Now().UnixNano())
	addr := "prisma-airs_supply_chain_skill_scanning_override.test"
	ids := map[string]bool{}
	config := func(reason string) string {
		return skillLiveConfig() + fmt.Sprintf(`resource "prisma-airs_supply_chain_skill_scanning_override" "test" {
 skill_name=%q
 fingerprint=%q
 trusted_by="terraform-e2e@example.invalid"
 reason=%q
}
data "prisma-airs_supply_chain_skill_scanning_overrides" "owned" {
 fingerprint=prisma-airs_supply_chain_skill_scanning_override.test.fingerprint
}
output "owned_override" {
 value=data.prisma-airs_supply_chain_skill_scanning_overrides.owned.result.skill_overrides[0].uuid
 sensitive=true
 }
`, name, fingerprint, reason)
	}
	check := func(st *terraform.State) error {
		id := st.RootModule().Resources[addr].Primary.ID
		ids[id] = true
		if out, ok := st.RootModule().Outputs["owned_override"]; !ok || out.Value != id {
			return fmt.Errorf("discovery did not find owned override")
		}
		fmt.Printf("SKILL_OVERRIDE_OWNED id=%s\n", id)
		return nil
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("initial disposable fixture"), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "fingerprint", fingerprint), check)},
		{ResourceName: addr, ImportState: true, ImportStateVerify: true},
		{Config: config("replacement disposable fixture"), Check: check},
		{Config: config("replacement disposable fixture"), PlanOnly: true},
	}, CheckDestroy: func(*terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		rows, e := skillLiveOverrides(ctx, c)
		if e != nil {
			return e
		}
		for _, row := range rows {
			if ids[row.UUID] || row.Fingerprint == fingerprint {
				return fmt.Errorf("owned trusted override still exists")
			}
		}
		fmt.Printf("SKILL_OVERRIDE_CLEANUP owned=%d confirmed_absent=true\n", len(ids))
		return nil
	}})
}
func TestAccSkillScanningPolicyNoopAndDiscovery(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("live test requires TF_ACC=1")
	}
	testAccPreCheck(t)
	c := skillLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	p, e := c.RuleInstances.List(ctx, ag.ListOpts{Limit: 100})
	if e != nil {
		t.Fatal(e)
	}
	if len(p.RuleInstances) == 0 {
		t.Skip("no effective rule available")
	}
	selected := p.RuleInstances[0]
	addr := "prisma-airs_supply_chain_skill_scanning_rule.test"
	config := skillLiveConfig() + fmt.Sprintf(`resource "prisma-airs_supply_chain_skill_scanning_rule" "test" {
 rule_uuid=%q
 state=%q
}
data "prisma-airs_supply_chain_skill_scanning_rules" "catalog" {}
data "prisma-airs_supply_chain_skill_scanning_rule_instances" "effective" {}`, selected.RuleUUID, selected.State)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "original_state", string(selected.State)), resource.TestCheckResourceAttr(addr, "id", selected.RuleUUID))},
		{Config: config, PlanOnly: true},
		{ResourceName: addr, ImportState: true, ImportStateVerify: true},
	}, CheckDestroy: func(*terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		p, e := c.RuleInstances.List(ctx, ag.ListOpts{Limit: 100})
		if e != nil {
			return e
		}
		for _, row := range p.RuleInstances {
			if row.RuleUUID == selected.RuleUUID && row.State == selected.State {
				fmt.Printf("SKILL_RULE_RESTORATION unchanged_policy=true rule=%s state=%s\n", row.RuleUUID, row.State)
				return nil
			}
		}
		return fmt.Errorf("adopted rule did not retain its original state")
	}})
}

func TestAccSkillScanningReadOnlyDiscovery(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("live test requires TF_ACC=1")
	}
	testAccPreCheck(t)
	c := skillLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	page, e := c.Scans.List(ctx, ag.ScanListOpts{ListOpts: ag.ListOpts{Limit: 100}, ScanFilter: ag.ScanFilter{Status: s.AgentGuardScanStatusCompleted}})
	if e != nil {
		t.Fatal(e)
	}
	if len(page.Scans) == 0 {
		t.Skip("no completed scan in the configured tenant")
	}
	scan := page.Scans[0]
	config := skillLiveConfig() + fmt.Sprintf(`
data "prisma-airs_supply_chain_skill_scanning_scan" "existing" { scan_uuid=%q }
data "prisma-airs_supply_chain_skill_scanning_scans" "existing" {
 limit=10
 artifact_types=["SKILL"]
}
data "prisma-airs_supply_chain_skill_scanning_vulnerabilities" "existing" {
 scan_uuid=%q
 in_chain=false
}
data "prisma-airs_supply_chain_skill_scanning_attack_chains" "existing" { scan_uuid=%q }
data "prisma-airs_supply_chain_skill_scanning_statistics" "existing" { time_period="7_DAYS" }
output "existing_scan_status" {
 value=data.prisma-airs_supply_chain_skill_scanning_scan.existing.result.status
 sensitive=true
}
`, scan.UUID, scan.UUID, scan.UUID)
	if fingerprint, ok := scan.Fingerprint.Get(); ok && len(fingerprint) == 64 {
		config += fmt.Sprintf(`data "prisma-airs_supply_chain_skill_scanning_scan" "lookup" { fingerprint=%q }
`, fingerprint)
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{{Config: config, Check: resource.TestCheckOutput("existing_scan_status", "COMPLETED")}, {Config: config, PlanOnly: true}}})
	fmt.Printf("SKILL_READ_DISCOVERY scan=%s completed=true\n", scan.UUID)
}
