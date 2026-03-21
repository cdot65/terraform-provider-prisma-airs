package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedTeamCategoriesDataSource_basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedTeamCategoriesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.prisma-airs_red_team_categories.test", "categories.#"),
				),
			},
		},
	})
}

func testAccRedTeamCategoriesConfig() string {
	return `
data "prisma-airs_red_team_categories" "test" {}
`
}
