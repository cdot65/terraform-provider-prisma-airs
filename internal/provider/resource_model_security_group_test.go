package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccModelSecurityGroupResource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := "tf-test-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelSecurityGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prisma-airs_model_security_group.test", "name", rName),
					resource.TestCheckResourceAttr("prisma-airs_model_security_group.test", "description", "test security group"),
					resource.TestCheckResourceAttrSet("prisma-airs_model_security_group.test", "id"),
					resource.TestCheckResourceAttrSet("prisma-airs_model_security_group.test", "uuid"),
				),
			},
			{
				ResourceName:            "prisma-airs_model_security_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state", "updated_at"},
			},
		},
	})
}

func testAccModelSecurityGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "prisma-airs_model_security_group" "test" {
  name        = %[1]q
  description = "test security group"
  source_type = "HUGGING_FACE"
}
`, name)
}
