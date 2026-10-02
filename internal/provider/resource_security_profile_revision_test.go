package provider_test

import (
	"context"
	"encoding/json"
	"fmt"
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"regexp"
	"strings"
	"testing"
)

func accProfileRevisions(ctx context.Context, client *airsruntime.Client, name string) ([]airsruntime.SecurityProfile, error) {
	latest := false
	var result []airsruntime.SecurityProfile
	for offset := 0; ; {
		page, err := client.Profiles.ListWithOptions(ctx, airsruntime.ProfileListOpts{ListOpts: airsruntime.ListOpts{Limit: 100, Offset: offset}, Latest: &latest})
		if err != nil {
			return nil, err
		}
		for _, p := range page.Items {
			if p.ProfileName == name {
				result = append(result, p)
			}
		}
		if page.NextOffset > offset {
			offset = page.NextOffset
		} else if len(page.Items) >= 100 {
			offset += len(page.Items)
		} else {
			return result, nil
		}
	}
}

type profileBeforeCheck struct {
	addr     string
	expected *string
}

func (c profileBeforeCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, change := range req.Plan.ResourceChanges {
		if change.Address == c.addr {
			before, ok := change.Change.Before.(map[string]any)
			if !ok || before["profile_id"] != *c.expected {
				resp.Error = fmt.Errorf("refresh did not follow highest remaining UUID")
			}
			return
		}
	}
	resp.Error = fmt.Errorf("profile missing from plan")
}
func TestAccSecurityProfileResource_revisionsAndRename(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	name := "tf-acc-rev-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	newName := name + "-new"
	const addr = "prisma-airs_security_profile.test"
	var current idRecorder
	var first, expectedBefore string
	t.Cleanup(func() {
		ctx, cancel := accCtx()
		defer cancel()
		for _, name := range []string{name, newName} {
			revs, err := accProfileRevisions(ctx, client, name)
			if err != nil {
				t.Error("fixture history cleanup list failed")
				continue
			}
			for _, p := range revs {
				if _, err := client.Profiles.ForceDelete(ctx, p.ProfileID, "terraform-acc"); err != nil && !gone404(err) {
					t.Errorf("fixture revision cleanup failed: %s", p.ProfileID)
				}
			}
			remaining, err := accProfileRevisions(ctx, client, name)
			if err != nil || len(remaining) > 0 {
				t.Errorf("fixture cleanup unconfirmed for %s", name)
			}
		}
	})
	updateChecks := resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate),
		profilePolicyDeltaCheck{addr, "allow"},
		plancheck.ExpectKnownValue(addr, tfjsonpath.New("ai_security_profile").AtSliceIndex(0).AtMapKey("model_type"), knownvalue.StringExact("default")),
		plancheck.ExpectKnownValue(addr, tfjsonpath.New("ai_security_profile").AtSliceIndex(0).AtMapKey("model_protection").AtSliceIndex(0).AtMapKey("name"), knownvalue.StringExact("prompt-injection")),
		plancheck.ExpectUnknownValue(addr, tfjsonpath.New("profile_id")), plancheck.ExpectUnknownValue(addr, tfjsonpath.New("revision")),
	}}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			ctx, cancel := accCtx()
			defer cancel()
			revs, err := accProfileRevisions(ctx, client, newName)
			if err != nil {
				return err
			}
			if len(revs) > 0 {
				return fmt.Errorf("current named profile history remains")
			}
			old, err := accProfileRevisions(ctx, client, name)
			if err != nil {
				return err
			}
			if len(old) < 2 {
				return fmt.Errorf("rename did not preserve prior named profile history")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{Config: testAccSecurityProfileConfig(name, "block"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "revision", "1"), current.capture("profile_id", addr), func(_ *terraform.State) error { first = current.value; return nil })},
			{Config: testAccSecurityProfileConfig(name, "allow"), ConfigPlanChecks: updateChecks, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "revision", "2"), current.capture("profile_id", addr), func(_ *terraform.State) error {
				if current.value == first {
					return fmt.Errorf("policy update did not create new UUID")
				}
				return nil
			})},
			{ResourceName: addr, ImportState: true, ImportStateId: name, ImportStateVerify: true},
			{PreConfig: func() {
				ctx, cancel := accCtx()
				defer cancel()
				revs, err := accProfileRevisions(ctx, client, name)
				if err != nil {
					t.Fatal(err)
				}
				var latest airsruntime.SecurityProfile
				for _, p := range revs {
					if p.Revision > latest.Revision {
						latest = p
					}
				}
				policy := latest.Policy
				policy.AiSecurityProfiles[0].ModelConfiguration.ModelProtection[0].Action = "block"
				created, err := client.Profiles.Update(ctx, latest.ProfileID, airsruntime.UpdateProfileRequest{ProfileName: name, Policy: policy})
				if err != nil {
					t.Fatal(err)
				}
				expectedBefore = created.ProfileID
			}, Config: testAccSecurityProfileConfig(name, "allow"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{profileBeforeCheck{addr, &expectedBefore}, plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)}}, Check: current.capture("profile_id", addr)},
			{PreConfig: func() {
				ctx, cancel := accCtx()
				defer cancel()
				revs, err := accProfileRevisions(ctx, client, name)
				if err != nil {
					t.Fatal(err)
				}
				var highest int32
				for _, p := range revs {
					if p.ProfileID != current.value && p.Revision > highest {
						highest = p.Revision
						expectedBefore = p.ProfileID
					}
				}
				if _, err := client.Profiles.ForceDelete(ctx, current.value, "terraform-acc"); err != nil {
					t.Fatal(err)
				}
			}, Config: testAccSecurityProfileConfig(name, "allow"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{profileBeforeCheck{addr, &expectedBefore}}}},
			{Config: testAccSecurityProfileConfig(newName, "allow"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate), plancheck.ExpectUnknownValue(addr, tfjsonpath.New("profile_id")), plancheck.ExpectUnknownValue(addr, tfjsonpath.New("revision"))}}, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "profile_name", newName), resource.TestCheckResourceAttr(addr, "revision", "1"))},
			{Config: testAccSecurityProfileConfig(newName, "block"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)}}, Check: resource.TestCheckResourceAttr(addr, "revision", "2")},
			{Config: testAccSecurityProfileConfig(name, "allow"), ExpectError: regexp.MustCompile("explicitly import")},
		}})
}

type profilePolicyDeltaCheck struct{ addr, action string }

func (c profilePolicyDeltaCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, change := range req.Plan.ResourceChanges {
		if change.Address == c.addr {
			before := change.Change.Before.(map[string]any)["ai_security_profile"]
			encoded, err := json.Marshal(before)
			if err != nil {
				resp.Error = err
				return
			}
			var expected []any
			if err = json.Unmarshal(encoded, &expected); err != nil {
				resp.Error = err
				return
			}
			expected[0].(map[string]any)["model_protection"].([]any)[0].(map[string]any)["action"] = c.action
			after := change.Change.After.(map[string]any)["ai_security_profile"]
			expectedJSON, expectedErr := json.Marshal(expected)
			afterJSON, afterErr := json.Marshal(after)
			// Normalize JSON number representations (json.Number versus float64)
			// while still comparing every configured/computed policy leaf.
			if expectedErr != nil || afterErr != nil || string(expectedJSON) != string(afterJSON) {
				resp.Error = fmt.Errorf("policy plan changed fields beyond the requested action")
			}
			beforeDLP, _ := json.Marshal(change.Change.Before.(map[string]any)["dlp_data_profile"])
			afterDLP, _ := json.Marshal(change.Change.After.(map[string]any)["dlp_data_profile"])
			if string(beforeDLP) != string(afterDLP) {
				resp.Error = fmt.Errorf("policy edit changed unrelated DLP configuration")
			}
			return
		}
	}
	resp.Error = fmt.Errorf("profile absent from policy delta plan")
}

// Rich policy regression: explicit false and omitted canonical defaults must
// survive apply/import and remain known during a one-leaf revision update.
func TestAccSecurityProfileResource_richPolicy(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	ctx, cancel := accCtx()
	defer cancel()
	dlps, err := client.DlpProfiles.List(ctx, airsruntime.ListOpts{Limit: 100})
	if err != nil || len(dlps.Items) == 0 {
		t.Fatal("a readable preexisting DLP profile is required")
	}
	dlp := dlps.Items[0]
	var second airsruntime.DlpProfile
	for _, candidate := range dlps.Items {
		if candidate.ID != "" && candidate.ID != dlp.ID {
			second = candidate
			break
		}
	}
	if second.ID == "" {
		t.Fatal("two distinct preexisting DLP profiles are required for reference-switch verification")
	}

	fmt.Printf("BORROWED fixture=dlp-profile id=%s read-only=true\n", dlp.ID)
	name := "tf-acc-rich-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	accProfileCleanup(t, client, name)
	const addr = "prisma-airs_security_profile.test"
	config := func(action string) string {
		return fmt.Sprintf(`resource "prisma-airs_security_profile" "test" {
 profile_name = %q
 ai_security_profile {
  latency { max_inline_latency = 5 }
  data_protection {
   data_leak_detection {
    action = "block"
    mask_data_inline = false
    member {
     text = %q
     id = %q
     version = %q
    }
   }
  }
  app_protection {
   default_url_category = ["malicious"]
   url_detected_action = "block"
   malicious_code_protection {
    name = "malicious-code"
    action = "block"
   }
  }
  model_protection {
   name = "prompt-injection"
   action = %q
  }
 }
 dlp_data_profile {
  log_severity = "low"
  profile_id = %q
  uuid = %q
  version = %q
 }
}`, name, dlp.Name, dlp.ID, dlp.Version, action, dlp.ID, dlp.UUID, dlp.Version)
	}
	namedReference := strings.Replace(config("allow"), "dlp_data_profile {", fmt.Sprintf("dlp_data_profile {\n name = %q\n file_based = \"block\"\n non_file_based = \"allow\"", dlp.Name), 1)
	switchedReference := config("allow")
	dlpStart := strings.Index(switchedReference, "dlp_data_profile {")
	suffix := switchedReference[dlpStart:]
	suffix = strings.Replace(suffix, fmt.Sprintf("profile_id = %q", dlp.ID), fmt.Sprintf("profile_id = %q", second.ID), 1)
	suffix = strings.Replace(suffix, fmt.Sprintf("uuid = %q", dlp.UUID), fmt.Sprintf("uuid = %q", second.UUID), 1)
	suffix = strings.Replace(suffix, fmt.Sprintf("version = %q", dlp.Version), fmt.Sprintf("version = %q", second.Version), 1)
	switchedReference = switchedReference[:dlpStart] + suffix

	check := resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "ai_security_profile.0.data_protection.data_leak_detection.mask_data_inline", "false"), resource.TestCheckResourceAttr(addr, "ai_security_profile.0.latency.max_inline_latency", "5"))
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_security_profile", "profile_name", func(ctx context.Context, name string) (bool, error) {
		revs, err := accProfileRevisions(ctx, client, name)
		return len(revs) == 0, err
	}), Steps: []resource.TestStep{
		{Config: config("block"), Check: check},
		{ResourceName: addr, ImportState: true, ImportStateId: name, ImportStateVerify: true},
		{Config: config("allow"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate), profilePolicyDeltaCheck{addr, "allow"}, plancheck.ExpectUnknownValue(addr, tfjsonpath.New("profile_id"))}}, Check: check},
		{Config: config("allow")},
		{Config: namedReference, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "dlp_data_profile.0.name", dlp.Name), resource.TestCheckResourceAttr(addr, "dlp_data_profile.0.file_based", "block"))},
		{Config: switchedReference, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate), plancheck.ExpectUnknownValue(addr, tfjsonpath.New("dlp_data_profile").AtSliceIndex(0).AtMapKey("name")), plancheck.ExpectUnknownValue(addr, tfjsonpath.New("dlp_data_profile").AtSliceIndex(0).AtMapKey("file_based"))}}, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "dlp_data_profile.0.profile_id", second.ID), resource.TestCheckResourceAttr(addr, "dlp_data_profile.0.name", ""), resource.TestCheckResourceAttr(addr, "dlp_data_profile.0.file_based", ""), resource.TestCheckResourceAttr(addr, "dlp_data_profile.0.non_file_based", ""))},
	}})
}
