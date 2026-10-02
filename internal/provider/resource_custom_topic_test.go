package provider_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCustomTopicResource_lifecycle(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_runtime_custom_topic.test"
	var rec idRecorder

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: destroyCheck("prisma-airs_runtime_custom_topic", "topic_id", func(ctx context.Context, id string) (bool, error) {
			for offset := 0; ; {
				page, err := client.Topics.List(ctx, airsruntime.ListOpts{Limit: 100, Offset: offset})
				if err != nil {
					return false, err
				}
				for _, it := range page.Items {
					if it.TopicID == id {
						return false, nil
					}
				}
				if page.NextOffset > offset {
					offset = page.NextOffset
				} else if len(page.Items) >= 100 {
					offset += len(page.Items)
				} else {
					return true, nil
				}
			}
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomTopicConfig(name, "test topic", `["example one", "example two"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "topic_name", name),
					resource.TestCheckResourceAttr(addr, "description", "test topic"),
					resource.TestCheckResourceAttr(addr, "examples.#", "2"),
					resource.TestCheckResourceAttrSet(addr, "topic_id"),
				),
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomTopicConfig(name, "updated topic", `["example one", "example two", "example three"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "description", "updated topic"),
					resource.TestCheckResourceAttr(addr, "examples.#", "3"),
					rec.capture("topic_id", addr),
				),
			},
			{Config: testAccCustomTopicConfig(name, "updated topic", `[]`), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "examples.#", "0"), rec.capture("topic_id", addr))},
			// Remote deletion: Read must drop the topic from state and the next apply recreates it.
			{
				PreConfig: func() {
					ctx, cancel := accCtx()
					defer cancel()
					if _, err := client.Topics.ForceDelete(ctx, rec.value, "terraform-acc"); err != nil {
						t.Fatalf("remote delete: %v", err)
					}
				},
				Config: testAccCustomTopicConfig(name, "updated topic", `["example one", "example two", "example three"]`),
				Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttrSet(addr, "topic_id"), rec.different("topic_id", addr)),
			},
		},
	})
}

func testAccCustomTopicConfig(name, description, examples string) string {
	return fmt.Sprintf(`
resource "prisma-airs_runtime_custom_topic" "test" {
  topic_name  = %[1]q
  description = %[2]q
  examples    = %[3]s
}
`, name, description, examples)
}
func TestAccCustomTopicResource_omittedDescription(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	name := "tf-acc-null-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var description string
	config := fmt.Sprintf(`resource "prisma-airs_runtime_custom_topic" "test" {
 topic_name = %q
 examples = []
}`, name)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_runtime_custom_topic", "topic_id", func(ctx context.Context, id string) (bool, error) {
		for offset := 0; ; {
			page, err := client.Topics.List(ctx, airsruntime.ListOpts{Limit: 100, Offset: offset})
			if err != nil {
				return false, err
			}
			for _, topic := range page.Items {
				if topic.TopicID == id {
					return false, nil
				}
			}
			if page.NextOffset > offset {
				offset = page.NextOffset
			} else if len(page.Items) >= 100 {
				offset += len(page.Items)
			} else {
				return true, nil
			}
		}
	}), Steps: []resource.TestStep{{Config: config, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttrSet("prisma-airs_runtime_custom_topic.test", "description"), resource.TestCheckResourceAttr("prisma-airs_runtime_custom_topic.test", "examples.#", "0"))}, {Config: config, Check: func(state *terraform.State) error {
		value := state.RootModule().Resources["prisma-airs_runtime_custom_topic.test"].Primary.Attributes["description"]
		if value == "" {
			return fmt.Errorf("description unavailable")
		}
		description = value
		return nil
	}}, {Config: strings.Replace(config, "examples = []", `examples = ["new example"]`, 1), Check: func(state *terraform.State) error {
		if state.RootModule().Resources["prisma-airs_runtime_custom_topic.test"].Primary.Attributes["description"] != description {
			return fmt.Errorf("omitted description changed while updating examples")
		}
		return nil
	}}}})
}
