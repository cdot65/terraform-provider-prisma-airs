package provider_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSecurityProfileResource_lifecycle(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_security_profile.test"
	var rec idRecorder
	accProfileCleanup(t, client, name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: destroyCheck("prisma-airs_security_profile", "profile_name", func(ctx context.Context, name string) (bool, error) {
			revisions, err := accProfileRevisions(ctx, client, name)
			return len(revisions) == 0, err
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileConfig(name, "block"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "profile_name", name),
					resource.TestCheckResourceAttrSet(addr, "profile_id"),
					resource.TestCheckResourceAttr(addr, "ai_security_profile.0.model_type", "default"),
					resource.TestCheckResourceAttr(addr, "ai_security_profile.0.model_protection.0.name", "prompt-injection"),
					resource.TestCheckResourceAttr(addr, "ai_security_profile.0.model_protection.0.action", "block"),
				),
			},
			{
				// Profiles are imported by name.
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[addr].Primary.Attributes["profile_name"], nil
				},
			},
			{
				Config: testAccSecurityProfileConfig(name, "allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "ai_security_profile.0.model_protection.0.action", "allow"),
					rec.capture("profile_id", addr),
				),
			},
			{
				PreConfig: func() {
					ctx, cancel := accCtx()
					defer cancel()
					if _, err := client.Profiles.ForceDelete(ctx, rec.value, "terraform-acc"); err != nil {
						t.Fatalf("remote delete: %v", err)
					}
				},
				Config: testAccSecurityProfileConfig(name, "allow"),
				Check:  resource.TestCheckResourceAttrSet(addr, "profile_id"),
			},
		},
	})
}

// TestAccSecurityProfileResource_topicReference proves topic_name is resolved to
// topic_id/revision on create and update without leaving a plan diff.
func TestAccSecurityProfileResource_topicReference(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_security_profile.test"

	accProfileCleanup(t, client, "tf-acc-"+suffix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: destroyCheck("prisma-airs_security_profile", "profile_name", func(ctx context.Context, name string) (bool, error) {
			revisions, err := accProfileRevisions(ctx, client, name)
			return len(revisions) == 0, err
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileTopicConfig(suffix, "block"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(addr, "ai_security_profile.0.model_protection.0.topic_list.0.topic.0.topic_id"),
					resource.TestCheckResourceAttrSet(addr, "ai_security_profile.0.model_protection.0.topic_list.0.topic.0.revision"),
				),
			},
			{
				Config: testAccSecurityProfileTopicConfig(suffix, "allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "ai_security_profile.0.model_protection.0.topic_list.0.action", "allow"),
					resource.TestCheckResourceAttrSet(addr, "ai_security_profile.0.model_protection.0.topic_list.0.topic.0.topic_id"),
				),
			},
		},
	})
}

func testAccSecurityProfileConfig(name, action string) string {
	return fmt.Sprintf(`
resource "prisma-airs_security_profile" "test" {
  profile_name = %[1]q

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = %[2]q
    }
  }
}
`, name, action)
}

func testAccSecurityProfileTopicConfig(suffix, action string) string {
	return fmt.Sprintf(`
resource "prisma-airs_custom_topic" "test" {
  topic_name  = "tf-acc-topic-%[1]s"
  description = "acceptance topic"
  examples    = ["acceptance example"]
}

resource "prisma-airs_security_profile" "test" {
  profile_name = "tf-acc-%[1]s"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "topic-guardrails"
      action = "block"

      topic_list {
        action = %[2]q

        topic {
          topic_name = prisma-airs_custom_topic.test.topic_name
        }
      }
    }
  }
}
`, suffix, action)
}
