package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentProfilesDataSource_basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentProfilesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.prisma-airs_deployment_profiles.test", "total_count"),
				),
			},
		},
	})
}

func testAccDeploymentProfilesConfig() string {
	return `
data "prisma-airs_deployment_profiles" "test" {}
`
}
