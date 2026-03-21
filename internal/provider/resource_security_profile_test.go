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
				),
			},
			// Note: ImportState step skipped — GET /v1/mgmt/profile?profile_name=X
			// returns 403. Read uses List-based lookup which works.
			// Update step skipped — PUT /v1/mgmt/profile/{id} returns 403
			// with the current service account permissions.
		},
	})
}

func testAccSecurityProfileConfig(name string) string {
	return fmt.Sprintf(`
resource "prisma-airs_security_profile" "test" {
  profile_name = %[1]q
  policy = jsonencode({
    "ai-security-profiles" = [
      {
        "model-type" = "default"
        "model-configuration" = {
          "model-protection" = [
            {
              "name"   = "prompt-injection"
              "action" = "alert"
            }
          ]
        }
      }
    ]
  })
}
`, name)
}
