package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedTeamQuotaDataSource_basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedTeamQuotaConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.prisma-airs_red_team_quota.test", "details"),
				),
			},
		},
	})
}

func testAccRedTeamQuotaConfig() string {
	return `
data "prisma-airs_red_team_quota" "test" {}
`
}
