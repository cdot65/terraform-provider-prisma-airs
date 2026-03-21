package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedTeamTargetResource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := "tf-test-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedTeamTargetConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prisma-airs_red_team_target.test", "name", rName),
					resource.TestCheckResourceAttr("prisma-airs_red_team_target.test", "description", "test target"),
					resource.TestCheckResourceAttrSet("prisma-airs_red_team_target.test", "id"),
					resource.TestCheckResourceAttrSet("prisma-airs_red_team_target.test", "uuid"),
				),
			},
			{
				ResourceName:            "prisma-airs_red_team_target.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connection_params"},
			},
		},
	})
}

func testAccRedTeamTargetConfig(name string) string {
	return fmt.Sprintf(`
resource "prisma-airs_red_team_target" "test" {
  name        = %[1]q
  description = "test target"
}
`, name)
}
