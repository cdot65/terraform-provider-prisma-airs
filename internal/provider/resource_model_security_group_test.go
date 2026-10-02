package provider_test

import (
	"context"
	"fmt"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccModelSecurityGroupResource_lifecycle(t *testing.T) {
	testAccPreCheck(t)
	client := accModelSecClient(t)
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_model_security_group.test"
	var rec idRecorder

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: destroyCheck("prisma-airs_model_security_group", "uuid", func(ctx context.Context, id string) (bool, error) {
			group, err := client.SecurityGroups.Get(ctx, id)
			if err == nil {
				return group.IsTombstone, nil
			}
			return goneOnNotFound(err)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccModelSecurityGroupConfig(name, "test security group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", name),
					resource.TestCheckResourceAttr(addr, "description", "test security group"),
					resource.TestCheckResourceAttrSet(addr, "uuid"),
				),
			},
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state", "updated_at"},
			},
			{
				Config: testAccModelSecurityGroupConfig(name+"-b", "updated security group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", name+"-b"),
					resource.TestCheckResourceAttr(addr, "description", "updated security group"),
					rec.capture("uuid", addr),
				),
			},
			{Config: testAccModelSecurityGroupConfig(name+"-b", ""), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)}}, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "description", ""), rec.capture("uuid", addr))},
			{
				PreConfig: func() {
					ctx, cancel := accCtx()
					defer cancel()
					if err := client.SecurityGroups.Delete(ctx, rec.value); err != nil {
						t.Fatalf("remote delete: %v", err)
					}
				},
				Config: testAccModelSecurityGroupConfig(name+"-b", "updated security group"),
				Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttrSet(addr, "uuid"), rec.different("uuid", addr)),
			},
		},
	})
}

func testAccModelSecurityGroupConfig(name, description string) string {
	return fmt.Sprintf(`
resource "prisma-airs_model_security_group" "test" {
  name        = %[1]q
  description = %[2]q
  source_type = "HUGGING_FACE"
}
`, name, description)
}

func TestAccModelSecurity_entitlement(t *testing.T) {
	testAccPreCheck(t)
	ctx, cancel := accCtx()
	defer cancel()
	if _, err := accModelSecClient(t).SecurityGroups.List(ctx, modelsecurity.GroupListOpts{Limit: 1}); err != nil {
		t.Fatal(err)
	}
}
func TestAccModelSecurityGroupResource_omittedDescriptionAndReplacement(t *testing.T) {
	testAccPreCheck(t)
	client := accModelSecClient(t)
	name := "tf-acc-null-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	config := func(source string) string {
		return fmt.Sprintf(`resource "prisma-airs_model_security_group" "test" {
 name = %q
 source_type = %q
}`, name, source)
	}
	const addr = "prisma-airs_model_security_group.test"
	var prior idRecorder
	checkPriorTombstone := func(_ *terraform.State) error {
		ctx, cancel := accCtx()
		defer cancel()
		group, err := client.SecurityGroups.Get(ctx, prior.value)
		if err != nil {
			return err
		}
		if !group.IsTombstone {
			return fmt.Errorf("prior source group remains active")
		}
		fmt.Printf("CLEANUP resource=prisma-airs_model_security_group id=%s outcome=tombstoned confirmed=true\n", prior.value)
		return nil
	}

	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_model_security_group", "uuid", func(ctx context.Context, id string) (bool, error) {
		group, err := client.SecurityGroups.Get(ctx, id)
		if err == nil {
			return group.IsTombstone, nil
		}
		return goneOnNotFound(err)
	}), Steps: []resource.TestStep{
		{Config: config("HUGGING_FACE"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "description", ""), prior.capture("uuid", addr))},
		{Config: config("LOCAL"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "source_type", "LOCAL"), prior.different("uuid", addr), checkPriorTombstone), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionReplace)}}},
	}})
}
