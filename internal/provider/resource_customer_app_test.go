package provider_test

import (
	"context"
	"fmt"
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"testing"
)

func TestAccCustomerAppResource_lifecycle(t *testing.T) {
	testAccPreCheck(t)
	client := accMgmtClient(t)
	code, creator := accKeyInputs(t, client)
	name := "tf-acc-app-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	ctx, cancel := accCtx()
	defer cancel()
	key, err := client.ApiKeys.Create(ctx, airsruntime.CreateApiKeyRequest{ApiKeyName: name, AuthCode: code, CustApp: name, CreatedBy: creator, CustEnv: "dev", CustCloudProvider: "aws", RotationTimeInterval: 1, RotationTimeUnit: "days"})
	if err != nil {
		t.Fatal("unable to create disposable key/app fixture")
	}
	if key.ApiKeyID == "" {
		t.Fatal("fixture create omitted key ID")
	}
	t.Cleanup(func() {
		ctx, cancel := accCtx()
		defer cancel()
		_, err := client.ApiKeys.Delete(ctx, name, creator)
		if err != nil && !gone404(err) {
			t.Error("disposable key cleanup failed")
		}
		_, err = accCustomerApp(ctx, client, name)
		gone, e := goneOnNotFound(err)
		if e != nil || !gone {
			t.Error("disposable application cleanup unconfirmed")
		}
	})
	const addr = "prisma-airs_customer_app.test"
	config := func(model string) string {
		configuredModel := fmt.Sprintf("model_name = %q", model)
		if model == "" {
			configuredModel = ""
		}
		return fmt.Sprintf(`resource "prisma-airs_customer_app" "test" {
 app_name = %q
 %s
 cloud_provider = "aws"
 environment = "dev"
}`, name, configuredModel)
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: destroyCheck("prisma-airs_customer_app", "app_name", func(ctx context.Context, name string) (bool, error) {
			_, err := accCustomerApp(ctx, client, name)
			return goneOnNotFound(err)
		}),
		Steps: []resource.TestStep{
			{Config: config(""), ResourceName: addr, ImportState: true, ImportStateId: name, ImportStatePersist: true},
			{Config: config("tf-acc-model"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "app_name", name), resource.TestCheckResourceAttr(addr, "model_name", "tf-acc-model"), resource.TestCheckResourceAttr(addr, "environment", "dev"), resource.TestCheckResourceAttr(addr, "cloud_provider", "aws"))},
			{ResourceName: addr, ImportState: true, ImportStateId: name, ImportStateVerify: true},
		}})
}
func gone404(err error) bool { gone, e := goneOnNotFound(err); return e == nil && gone }

func TestAccCustomerAppResource_remoteDeletion(t *testing.T) {
	// The lifecycle above verifies import/update/delete. This separate case proves
	// a missing imported app plans creation, which remains intentionally unsupported.
	testAccPreCheck(t)
	client := accMgmtClient(t)
	code, creator := accKeyInputs(t, client)
	name := "tf-acc-gone-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	ctx, cancel := accCtx()
	defer cancel()
	_, err := client.ApiKeys.Create(ctx, airsruntime.CreateApiKeyRequest{ApiKeyName: name, AuthCode: code, CustApp: name, CreatedBy: creator, CustEnv: "dev", CustCloudProvider: "aws", RotationTimeInterval: 1, RotationTimeUnit: "days"})
	if err != nil {
		t.Fatal("unable to create disposable app fixture")
	}
	t.Cleanup(func() {
		ctx, cancel := accCtx()
		defer cancel()
		_, err := client.ApiKeys.Delete(ctx, name, creator)
		if err != nil && !gone404(err) {
			t.Error("key cleanup failed")
		}
	})
	const addr = "prisma-airs_customer_app.test"
	config := fmt.Sprintf(`resource "prisma-airs_customer_app" "test" {app_name = %q}`, name)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, ResourceName: addr, ImportState: true, ImportStateId: name, ImportStatePersist: true},
		{PreConfig: func() {
			ctx, cancel := accCtx()
			defer cancel()
			_, err := client.CustomerApps.Delete(ctx, name, creator)
			if err != nil {
				_, readErr := accCustomerApp(ctx, client, name)
				if !gone404(readErr) {
					t.Fatal("remote app delete unconfirmed")
				}
			}
		}, RefreshState: true, ExpectNonEmptyPlan: true},
		{Config: config, PlanOnly: true, ExpectNonEmptyPlan: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionCreate)}}},
	}})
}
