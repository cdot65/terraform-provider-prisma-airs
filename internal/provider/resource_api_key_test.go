package provider_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"os"
	"strings"
	"testing"
)

func accKeyInputs(t *testing.T, client *airsruntime.Client) (string, string) {
	t.Helper()
	ctx, cancel := accCtx()
	defer cancel()
	profiles, err := client.DeploymentProfiles.List(ctx, airsruntime.ListOpts{})
	if err != nil {
		t.Fatalf("deployment profiles unavailable: %v", err)
	}
	code := ""
	profileName := ""
	for _, p := range profiles.Items {
		if p.AuthCode != "" {
			code = p.AuthCode
			profileName = p.DpName
			break
		}
	}
	if code == "" {
		t.Fatal("no preexisting deployment profile auth code available")
	}
	creator := os.Getenv("AIRS_ACC_CREATED_BY")
	if creator == "" {
		creator = "terraform-acc@example.com"
	}
	keys, err := client.ApiKeys.List(ctx, airsruntime.ListOpts{Limit: 1})
	if err != nil {
		t.Fatal("cannot read key fixture metadata")
	}
	if os.Getenv("AIRS_ACC_CREATED_BY") == "" && len(keys.Items) > 0 && keys.Items[0].CreatedBy != "" {
		creator = keys.Items[0].CreatedBy
	}
	fmt.Printf("BORROWED fixture=deployment-profile name=%q read-only=true creator_identity_sha256=%x\n", profileName, sha256.Sum256([]byte(creator)))
	return code, creator
}
func testAccApiKeyConfig(name, app, code, creator string, interval int) string {
	return fmt.Sprintf(`resource "prisma-airs_api_key" "test" {
 api_key_name = %q
 auth_code = %q
 cust_app = %q
 created_by = %q
 cust_env = "dev"
 cust_cloud_provider = "aws"
 rotation_time_interval = %d
 rotation_time_unit = "days"
}`, name, code, app, creator, interval)
}
func TestAccApiKeyResource_lifecycle(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	code, creator := accKeyInputs(t, client)
	name := "tf-acc-key-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_api_key.test"
	var secret string
	config := testAccApiKeyConfig(name, name, code, creator, 1)
	checkSecret := func(s *terraform.State) error {
		value := s.RootModule().Resources[addr].Primary.Attributes["api_key"]
		if value == "" {
			return fmt.Errorf("creation-time secret unavailable")
		}
		if secret != "" && secret != value {
			return fmt.Errorf("refresh changed creation-time secret")
		}
		secret = value
		return nil
	}
	steps := []resource.TestStep{
		{Config: config, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "api_key_name", name), resource.TestCheckResourceAttrSet(addr, "api_key_id"), checkSecret)},
		{Config: config, Check: checkSecret},
		{ResourceName: addr, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"api_key"}, ImportStateCheck: func(states []*terraform.InstanceState) error {
			if len(states) != 1 || states[0].Attributes["api_key"] != "" {
				return fmt.Errorf("import must leave one-time secret unavailable")
			}
			return nil
		}},
	}
	for _, mutation := range [][2]string{
		{fmt.Sprintf("api_key_name = %q", name), fmt.Sprintf("api_key_name = %q", name+"-b")},
		{fmt.Sprintf("auth_code = %q", code), `auth_code = "fixture-other-code"`},
		{fmt.Sprintf("cust_app = %q", name), fmt.Sprintf("cust_app = %q", name+"-b")},
		{fmt.Sprintf("created_by = %q", creator), `created_by = "fixture-other-creator"`},
		{`cust_env = "dev"`, `cust_env = "prod"`},
		{`cust_cloud_provider = "aws"`, `cust_cloud_provider = "gcp"`},
		{`rotation_time_interval = 1`, `rotation_time_interval = 2`},
		{`rotation_time_unit = "days"`, `rotation_time_unit = "months"`},
		{`rotation_time_unit = "days"`, `rotation_time_unit = "days"
 cust_ai_agent_framework = "fixture-framework"`},
	} {
		changed := strings.Replace(config, mutation[0], mutation[1], 1)
		if changed == config {
			t.Fatal("replacement mutation did not alter configuration")
		}
		steps = append(steps, resource.TestStep{Config: changed, PlanOnly: true, ExpectNonEmptyPlan: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionReplace)}}})
	}
	steps = append(steps, resource.TestStep{Config: config + "\n", Check: checkSecret})
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			ctx, cancel := accCtx()
			defer cancel()
			if err := destroyCheck("prisma-airs_api_key", "api_key_id", func(ctx context.Context, id string) (bool, error) {
				for offset := 0; ; {
					page, err := client.ApiKeys.List(ctx, airsruntime.ListOpts{Limit: 100, Offset: offset})
					if err != nil {
						return false, err
					}
					for _, key := range page.Items {
						if key.ApiKeyID == id {
							return false, nil
						}
					}
					if page.NextOffset > offset {
						offset = page.NextOffset
					} else if len(page.Items) >= 100 {
						offset += len(page.Items)
					} else {
						return true, nil
					}
				}
			})(s); err != nil {
				return err
			}
			_, err := accCustomerApp(ctx, client, name)
			gone, err := goneOnNotFound(err)
			if err != nil {
				return err
			}
			if !gone {
				return fmt.Errorf("associated disposable app remains after key deletion")
			}
			return nil
		},
		Steps: steps,
	})
}
