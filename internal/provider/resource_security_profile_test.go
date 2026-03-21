package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecurityProfileResource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := "tf-test-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prisma-airs_security_profile.test", "profile_name", rName),
					resource.TestCheckResourceAttrSet("prisma-airs_security_profile.test", "id"),
					resource.TestCheckResourceAttrSet("prisma-airs_security_profile.test", "profile_id"),
					resource.TestCheckResourceAttr("prisma-airs_security_profile.test", "ai_security_profile.0.model_type", "default"),
					resource.TestCheckResourceAttr("prisma-airs_security_profile.test", "ai_security_profile.0.model_protection.0.name", "prompt-injection"),
					resource.TestCheckResourceAttr("prisma-airs_security_profile.test", "ai_security_profile.0.model_protection.0.action", "alert"),
				),
			},
		},
	})
}

func testAccSecurityProfileConfig(name string) string {
	return fmt.Sprintf(`
resource "prisma-airs_security_profile" "test" {
  profile_name = %[1]q

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "alert"
    }
  }
}
`, name)
}
