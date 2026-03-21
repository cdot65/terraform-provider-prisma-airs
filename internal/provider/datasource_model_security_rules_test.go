package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccModelSecurityRulesDataSource_basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelSecurityRulesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.prisma-airs_model_security_rules.test", "rules.#"),
				),
			},
		},
	})
}

func testAccModelSecurityRulesConfig() string {
	return `
data "prisma-airs_model_security_rules" "test" {}
`
}
