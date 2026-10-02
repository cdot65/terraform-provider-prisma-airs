package provider_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// Prompt sets have no delete endpoint: destroy archives them. The test asserts
// the archived end state, and that an out-of-band archive drops the resource
// from state. The legacy `properties` metadata is deliberately not exercised
// because current service support for it is not guaranteed.
func TestAccRedTeamCustomPromptSetResource_lifecycle(t *testing.T) {
	testAccPreCheck(t)
	client := accRedTeamClient(t)
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_red_team_custom_prompt_set.test"
	var rec idRecorder

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: destroyCheck("prisma-airs_red_team_custom_prompt_set", "uuid", func(ctx context.Context, id string) (bool, error) {
			ps, err := client.CustomAttacks.GetPromptSet(ctx, id)
			if gone, gerr := goneOnNotFound(err); gone || gerr != nil {
				return gone, gerr
			}
			return ps.Archive, nil
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptSetConfig(name, "acceptance prompt set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", name),
					resource.TestCheckResourceAttr(addr, "archive", "false"),
					resource.TestCheckResourceAttrSet(addr, "uuid"),
				),
			},
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccPromptSetConfig(name, "updated prompt set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "description", "updated prompt set"),
					rec.capture("uuid", addr),
				),
			},
			{Config: testAccPromptSetConfig(name, ""), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionReplace)}}, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "description", ""), rec.capture("uuid", addr))},
			{
				// Archived out-of-band: Read removes it from state, apply creates a new set.
				PreConfig: func() {
					ctx, cancel := accCtx()
					defer cancel()
					if _, err := client.CustomAttacks.ArchivePromptSet(ctx, rec.value, redteamArchive()); err != nil {
						t.Fatalf("remote archive: %v", err)
					}
				},
				Config: testAccPromptSetConfig(name, "updated prompt set"),
				Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttrSet(addr, "uuid"), rec.different("uuid", addr)),
			},
		},
	})
}

func testAccPromptSetConfig(name, description string) string {
	return fmt.Sprintf(`
resource "prisma-airs_red_team_custom_prompt_set" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}
func TestAccRedTeamCustomPromptSetResource_omittedDescription(t *testing.T) {
	testAccPreCheck(t)
	client := accRedTeamClient(t)
	name := "tf-acc-null-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	config := fmt.Sprintf(`resource "prisma-airs_red_team_custom_prompt_set" "test" {name = %q}`, name)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_red_team_custom_prompt_set", "uuid", func(ctx context.Context, id string) (bool, error) {
		ps, err := client.CustomAttacks.GetPromptSet(ctx, id)
		if err != nil {
			return goneOnNotFound(err)
		}
		return ps.Archive, nil
	}), Steps: []resource.TestStep{{Config: config, Check: resource.TestCheckResourceAttr("prisma-airs_red_team_custom_prompt_set.test", "description", "")}, {Config: config}}})
}
