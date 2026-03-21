package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomTopicResource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := "tf-test-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomTopicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prisma-airs_custom_topic.test", "topic_name", rName),
					resource.TestCheckResourceAttr("prisma-airs_custom_topic.test", "description", "test topic"),
					resource.TestCheckResourceAttrSet("prisma-airs_custom_topic.test", "id"),
					resource.TestCheckResourceAttrSet("prisma-airs_custom_topic.test", "topic_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "prisma-airs_custom_topic.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Note: Update step skipped — the API returns 403 for PUT /v1/mgmt/topic/{id}
			// with the current service account permissions.
		},
	})
}

func testAccCustomTopicConfig(name string) string {
	return fmt.Sprintf(`
resource "prisma-airs_custom_topic" "test" {
  topic_name  = %[1]q
  description = "test topic"
  examples    = ["example one", "example two"]
}
`, name)
}
