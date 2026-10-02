package provider_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRedTeamTargetResource_lifecycle(t *testing.T) {
	testAccPreCheck(t)
	client := accRedTeamClient(t)
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_red_team_target.test"
	var rec idRecorder

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: destroyCheck("prisma-airs_red_team_target", "uuid", func(ctx context.Context, id string) (bool, error) {
			_, err := client.Targets.Get(ctx, id)
			return goneOnNotFound(err)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccRedTeamTargetConfig(name, "acceptance target"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", name),
					resource.TestCheckResourceAttr(addr, "target_type", "APPLICATION"),
					resource.TestCheckResourceAttrSet(addr, "uuid"),
					resource.TestCheckResourceAttrSet(addr, "status"),
				),
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
				// Arbitrary desired payloads and headers are unavailable on import.
				ImportStateVerifyIgnore: []string{"custom.request_body", "custom.response_body", "custom.request_headers"},
			},
			{
				Config: testAccRedTeamTargetConfig(name, "updated acceptance target"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "description", "updated acceptance target"),
					rec.capture("uuid", addr),
				),
			},
			{
				PreConfig: func() {
					ctx, cancel := accCtx()
					defer cancel()
					if _, err := client.Targets.Delete(ctx, rec.value); err != nil {
						t.Fatalf("remote delete: %v", err)
					}
				},
				Config: testAccRedTeamTargetConfig(name, "updated acceptance target"),
				Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttrSet(addr, "uuid"), rec.different("uuid", addr)),
			},
		},
	})
}

func testAccRedTeamTargetConfig(name, description string) string {
	return fmt.Sprintf(`
resource "prisma-airs_red_team_target" "test" {
  name            = %[1]q
  description     = %[2]q
  target_type     = "APPLICATION"
  custom {
 api_endpoint = "https://httpbin.org/post"
    request_headers = { "Content-Type" = "application/json" }
    request_body = { prompt = "{INPUT}" }
    response_body = { output = "{RESPONSE}" }
    response_key = "output"
  }
}
`, name, description)
}

func TestAccRedTeamTargetResource_nativeFamilies(t *testing.T) {
	testAccPreCheck(t)
	cases := map[string]string{
		"openai": `openai {
 api_key = "fixture-placeholder"
 model_name = "gpt-4o-mini"
 request_body = {prompt = "{INPUT}", flag = false, count = 0, omitted = null, items = []}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		"hugging_face": `hugging_face {
 api_key = "fixture-placeholder"
 model_name = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		"bedrock": `bedrock {
 access_id = "AKIAFIXTURE0000000000"
 access_secret = "fixture0000000000000000000000000000000000"
 region = "us-east-1"
 model_id = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		"databricks": `databricks {
 access_token = "fixture-placeholder"
 workspace_url = "https://example.cloud.databricks.com"
 model_name = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 response_stop_key = "done"
 response_stop_value = "true"
 }`,
		"databricks_oauth": `databricks {
 client_id = "fixture-client"
 secret = "fixture-secret"
 workspace_url = "https://example.cloud.databricks.com"
 model_name = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 response_stop_key = "done"
 response_stop_value = "true"
 }`,
		"custom": `custom {
 api_endpoint = "https://httpbin.org/post"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		"rest": `rest {
 api_endpoint = "https://httpbin.org/post"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		"streaming": `streaming {
 api_endpoint = "https://httpbin.org/post"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 response_stop_key = "done"
 response_stop_value = "true"
 }`,
	}
	for name, block := range cases {
		t.Run(name, func(t *testing.T) {
			client := accRedTeamClient(t)
			resourceName := "tf-acc-native-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
			config := func(description string) string {
				return fmt.Sprintf(`resource "prisma-airs_red_team_target" "test" {
 name = %q
 description = %q
 %s
}`, resourceName, description, block)
			}
			family := strings.TrimSuffix(name, "_oauth")
			const addr = "prisma-airs_red_team_target.test"
			checks := []resource.TestCheckFunc{resource.TestCheckResourceAttr(addr, "name", resourceName), resource.TestCheckResourceAttrSet(addr, "uuid")}
			if family == "openai" || family == "hugging_face" || family == "databricks" {
				checks = append(checks, resource.TestCheckResourceAttr(addr, family+".model_name", map[string]string{"openai": "gpt-4o-mini", "hugging_face": "fixture-model", "databricks": "fixture-model"}[family]))
			}
			ignore := []string{family + ".request_body", family + ".response_body"}
			for _, secret := range map[string][]string{"openai": {"api_key"}, "hugging_face": {"api_key"}, "databricks": {"access_token", "client_id", "secret"}, "bedrock": {"access_id", "access_secret"}}[family] {
				ignore = append(ignore, family+"."+secret)
			}
			if family == "openai" || family == "hugging_face" || family == "bedrock" || family == "databricks" {
				ignore = append(ignore, family+".response_key")
			}
			var prior idRecorder
			steps := []resource.TestStep{
				{Config: config("native fixture"), Check: resource.ComposeAggregateTestCheckFunc(checks...)},
				{ResourceName: addr, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: ignore, ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return family + "/" + state.RootModule().Resources[addr].Primary.ID, nil
				}},
				{Config: config("updated fixture"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)}}, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "description", "updated fixture"), prior.capture("uuid", addr))},
			}
			if family == "openai" || family == "hugging_face" || family == "databricks" || family == "bedrock" {
				changed := strings.NewReplacer("fixture-placeholder", "fixture-rotated", "fixture-secret", "fixture-secret-rotated", "fixture0000000000000000000000000000000000", "rotated0000000000000000000000000000000000", "gpt-4o-mini", "gpt-4o", "fixture-model", "fixture-model-next", "us-east-1", "us-west-2").Replace(config("updated fixture"))
				steps = append(steps, resource.TestStep{Config: changed, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionReplace)}}, Check: prior.different("uuid", addr)})
			}
			resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_red_team_target", "uuid", func(ctx context.Context, id string) (bool, error) {
				_, err := client.Targets.GetDetails(ctx, id)
				return goneOnNotFound(err)
			}), Steps: steps})
		})
	}
}
func TestAccRedTeamTargetResource_authenticationAndTransitions(t *testing.T) {
	testAccPreCheck(t)
	client := accRedTeamClient(t)
	name := "tf-acc-auth-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	auth := []string{`headers_auth {headers = {Authorization = "Bearer fixture-placeholder"}}`, `basic_auth {
 username = "fixture-user"
 password = "fixture-placeholder"
 }`, `oauth2_auth {
 token_url = "https://example.com/token"
 body = {client_id = "fixture-client", client_secret = "fixture-placeholder"}
 headers = {"Content-Type" = "application/json"}
 inject_header = {Authorization = "Bearer {TOKEN}"}
 response_key = "access_token"
 expiry_minutes = 60
 }`, ""}
	const addr = "prisma-airs_red_team_target.test"
	var steps []resource.TestStep
	for index, authentication := range auth {
		steps = append(steps, resource.TestStep{Config: fmt.Sprintf(`resource "prisma-airs_red_team_target" "test" {
 name = %q
 rest {
 api_endpoint = "https://httpbin.org/post"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }
 %s
}`, name, authentication), Check: resource.TestCheckResourceAttr(addr, "name", name)})
		if index > 0 {
			action := plancheck.ResourceActionUpdate
			if authentication == "" {
				action = plancheck.ResourceActionReplace
			}
			steps[len(steps)-1].ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, action)}}
		}
		if index == 2 {
			steps[len(steps)-1].Config = strings.Replace(steps[len(steps)-1].Config, ` response_key = "access_token"`+"\n", "", 1)
			steps[len(steps)-1].Config = strings.Replace(steps[len(steps)-1].Config, " expiry_minutes = 60\n", "", 1)
			steps[len(steps)-1].Check = resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "oauth2_auth.response_key", "access_token"), resource.TestCheckResourceAttr(addr, "oauth2_auth.expiry_minutes", "60"))
		}
	}
	steps = append(steps, resource.TestStep{Config: fmt.Sprintf(`resource "prisma-airs_red_team_target" "test" {
 name = %q
 streaming {
 api_endpoint = "https://httpbin.org/post"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 response_stop_key = "done"
 response_stop_value = "true"
 }
}`, name), Check: resource.TestCheckResourceAttr(addr, "response_mode", "STREAMING")})
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_red_team_target", "uuid", func(ctx context.Context, id string) (bool, error) {
		_, err := client.Targets.GetDetails(ctx, id)
		return goneOnNotFound(err)
	}), Steps: steps})
}

func TestAccRedTeamTargetResource_networkBrokerRemoval(t *testing.T) {
	testAccPreCheck(t)
	client := accRedTeamClient(t)
	channel := os.Getenv("AIRS_ACC_NETWORK_BROKER_CHANNEL_UUID")
	if channel == "" {
		t.Fatal("AIRS_ACC_NETWORK_BROKER_CHANNEL_UUID must identify a preexisting read-only channel")
	}
	ctx, cancel := accCtx()
	defer cancel()
	if _, err := client.NetworkBroker.Get(ctx, channel); err != nil {
		t.Fatal("borrowed channel lookup failed")
	}
	fmt.Printf("BORROWED fixture=network-broker-channel uuid=%s read-only=true\n", channel)
	name := "tf-acc-nb-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_red_team_target.test"
	var prior idRecorder
	config := func(broker bool) string {
		cfg := testAccRedTeamTargetConfig(name, "borrowed channel fixture")
		if broker {
			cfg = strings.Replace(cfg, "  custom {", fmt.Sprintf("  api_endpoint_type = \"NETWORK_BROKER\"\n  network_broker_channel_uuid = %q\n  custom {", channel), 1)
		}
		return cfg
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_red_team_target", "uuid", func(ctx context.Context, id string) (bool, error) {
		_, err := client.Targets.GetDetails(ctx, id)
		return goneOnNotFound(err)
	}), Steps: []resource.TestStep{
		{Config: config(true), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "network_broker_channel_uuid", channel), prior.capture("uuid", addr))},
		{Config: config(false), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionReplace)}}, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "api_endpoint_type", "PUBLIC"), resource.TestCheckNoResourceAttr(addr, "network_broker_channel_uuid"), prior.different("uuid", addr))},
	}})
}

func TestAccRedTeamTargetResource_replacementTransitions(t *testing.T) {
	testAccPreCheck(t)
	client := accRedTeamClient(t)
	name := "tf-acc-shape-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	const addr = "prisma-airs_red_team_target.test"
	var prior idRecorder
	blocks := []string{
		`openai {
 api_key = "fixture-token"
 model_name = "gpt-4o-mini"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		`bedrock {
 access_id = "AKIAFIXTURE0000000000"
 access_secret = "fixture0000000000000000000000000000000000"
 session_token = "fixture-session"
 region = "us-east-1"
 model_id = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		`bedrock {
 access_id = "AKIAFIXTURE0000000000"
 access_secret = "fixture0000000000000000000000000000000000"
 region = "us-east-1"
 model_id = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 }`,
		`databricks {
 access_token = "fixture-token"
 workspace_url = "https://example.cloud.databricks.com"
 model_name = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 response_stop_key = "done"
 response_stop_value = "true"
 }`,
		`databricks {
 client_id = "fixture-client"
 secret = "fixture-secret"
 workspace_url = "https://example.cloud.databricks.com"
 model_name = "fixture-model"
 request_body = {prompt = "{INPUT}"}
 response_body = {output = "{RESPONSE}"}
 response_key = "output"
 response_stop_key = "done"
 response_stop_value = "true"
 }`,
	}
	var steps []resource.TestStep
	for i, block := range blocks {
		checks := []resource.TestCheckFunc{resource.TestCheckResourceAttrSet(addr, "uuid")}
		if i > 0 {
			checks = append(checks, prior.different("uuid", addr))
		}
		checks = append(checks, prior.capture("uuid", addr))
		step := resource.TestStep{Config: fmt.Sprintf("resource \"prisma-airs_red_team_target\" \"test\" {\n name = %q\n %s\n}", name, block), Check: resource.ComposeAggregateTestCheckFunc(checks...)}
		if i > 0 {
			step.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionReplace)}}
		}
		steps = append(steps, step)
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: destroyCheck("prisma-airs_red_team_target", "uuid", func(ctx context.Context, id string) (bool, error) {
		_, err := client.Targets.GetDetails(ctx, id)
		return goneOnNotFound(err)
	}), Steps: steps})
}
