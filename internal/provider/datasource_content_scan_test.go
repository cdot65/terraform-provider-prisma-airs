package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckScan(t *testing.T) {
	t.Helper()
	if os.Getenv("PANW_AI_SEC_API_KEY") == "" {
		t.Skip("PANW_AI_SEC_API_KEY not set")
	}
}

func TestAccContentScanDataSource_basic(t *testing.T) {
	testAccPreCheckScan(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContentScanConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.prisma-airs_content_scan.test", "category"),
					resource.TestCheckResourceAttrSet("data.prisma-airs_content_scan.test", "action"),
					resource.TestCheckResourceAttrSet("data.prisma-airs_content_scan.test", "scan_id"),
					resource.TestCheckResourceAttrSet("data.prisma-airs_content_scan.test", "report_id"),
				),
			},
		},
	})
}

func testAccContentScanConfig() string {
	profileName := os.Getenv("PANW_AI_SEC_PROFILE_NAME")
	if profileName == "" {
		profileName = "best-practice"
	}
	return fmt.Sprintf(`
data "prisma-airs_content_scan" "test" {
  prompt   = "What is the capital of France?"
  response = "The capital of France is Paris."
  profile_name = %q
}
`, profileName)
}
