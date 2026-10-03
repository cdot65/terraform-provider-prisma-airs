package provider_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	s "github.com/cdot65/prisma-airs-go/aisec/gateway/schema"
	gatewayproduct "github.com/cdot65/prisma-airs-provider/internal/products/gateway"
	fr "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func accGatewayClient(t *testing.T) *gw.Client {
	t.Helper()
	c, e := gw.NewClient(gw.Opts{ClientID: os.Getenv("PANW_MGMT_CLIENT_ID"), ClientSecret: os.Getenv("PANW_MGMT_CLIENT_SECRET"), TsgID: os.Getenv("PANW_MGMT_TSG_ID"), DataEndpoint: os.Getenv("PANW_AI_GW_DATA_ENDPOINT"), AdminEndpoint: os.Getenv("PANW_AI_GW_ADMIN_ENDPOINT"), TokenEndpoint: os.Getenv("PANW_MGMT_TOKEN_ENDPOINT")})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func accGatewayWorkspace(t *testing.T) string {
	t.Helper()
	w := os.Getenv("PANW_AI_GW_TEST_WORKSPACE_ID")
	if w == "" {
		t.Fatal("live Gateway tests require an existing PANW_AI_GW_TEST_WORKSPACE_ID")
	}
	return w
}
func gatewayDoc(v any, e error) (map[string]any, error) {
	if e != nil {
		return nil, e
	}
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	var d map[string]any
	e = json.Unmarshal(b, &d)
	return d, e
}
func gatewayRead(ctx context.Context, c *gw.Client, kind, id, w string) (map[string]any, error) {
	switch kind {
	case "config":
		return gatewayDoc(c.Configs.Get(ctx, id))
	case "guardrail":
		return gatewayDoc(c.Guardrails.Get(ctx, id))
	case "org_guardrail":
		return gatewayDoc(c.OrgGuardrails.Get(ctx, id))
	case "integration":
		return gatewayDoc(c.Integrations.Get(ctx, id))
	case "provider":
		return gatewayDoc(c.Providers.Get(ctx, id, s.ProvidersGetOptions{WorkspaceID: &w}))
	case "mcp_integration":
		return gatewayDoc(c.MCPIntegrations.Get(ctx, id))
	case "mcp_server":
		return gatewayDoc(c.MCPServers.Get(ctx, id))
	case "service_api_key":
		return gatewayDoc(c.APIKeys.GetForKind(ctx, gw.APIKeyService, id))
	case "user_api_key":
		return gatewayDoc(c.APIKeys.GetForKind(ctx, gw.APIKeyUser, id))
	case "usage_limit":
		return gatewayDoc(c.UsageLimits.Get(ctx, id, s.UsageLimitsGetOptions{}))
	case "rate_limit":
		return gatewayDoc(c.RateLimits.Get(ctx, id, s.RateLimitsGetOptions{}))
	case "secret_reference":
		return gatewayDoc(c.SecretReferences.Get(ctx, id))
	case "deployment":
		return gatewayDoc(c.Deployments.Get(ctx, id))
	case "integration_workspace_binding", "mcp_integration_workspace_binding":
		parts := strings.Split(id, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid binding ID")
		}
		var v any
		var e error
		if kind == "integration_workspace_binding" {
			if _, err := c.Integrations.Get(ctx, parts[0]); err != nil {
				return nil, err
			}
		} else {
			if _, err := c.MCPIntegrations.Get(ctx, parts[0]); err != nil {
				return nil, err
			}
		}
		if kind == "integration_workspace_binding" {
			v, e = c.Integrations.ListWorkspaces(ctx, parts[0])
		} else {
			v, e = c.MCPIntegrations.ListWorkspaces(ctx, parts[0], s.MCPIntegrationsListWorkspacesOptions{})
		}
		if aisec.IsNotFound(e) {
			return nil, e
		}
		if e != nil {
			return nil, e
		}
		b, e := json.Marshal(v)
		if e != nil {
			return nil, e
		}
		var d map[string]any
		if e = json.Unmarshal(b, &d); e != nil {
			return nil, e
		}
		for _, key := range []string{"data", "workspaces"} {
			if list, ok := d[key].([]any); ok {
				for _, x := range list {
					item, ok := x.(map[string]any)
					if ok && item["id"] == parts[1] && item["enabled"] == true {
						return item, nil
					}
				}
			}
		}
		return nil, aisec.NewHTTPError("binding absent", aisec.ClientSideError, 404)
	}
	return nil, fmt.Errorf("unknown Gateway resource")
}
func gatewayDestroy(c *gw.Client, w string) func(*terraform.State) error {
	managed := map[string]bool{}
	for _, entry := range gatewayproduct.Definition().Resources {
		var metadata fr.MetadataResponse
		entry.New().Metadata(context.Background(), fr.MetadataRequest{ProviderTypeName: "prisma-airs"}, &metadata)
		managed[metadata.TypeName] = true
	}
	return func(st *terraform.State) error {
		ctx, cancel := accCtx()
		defer cancel()
		for _, r := range st.RootModule().Resources {
			if !managed[r.Type] || r.Primary == nil || r.Primary.ID == "" {
				continue
			}
			kind := strings.TrimPrefix(r.Type, "prisma-airs_gateway_")
			d, e := gatewayRead(ctx, c, kind, r.Primary.ID, w)
			if aisec.IsNotFound(e) {
				fmt.Printf("GATEWAY CLEANUP resource=%s id=%s outcome=absent confirmed=true\n", kind, r.Primary.ID)
				continue
			}
			if e != nil {
				return fmt.Errorf("Gateway destroy verification %s failed", kind)
			}
			if d["status"] == "archived" || d["status"] == "deleted" {
				fmt.Printf("GATEWAY CLEANUP resource=%s id=%s outcome=archived confirmed=true\n", kind, r.Primary.ID)
				continue
			}
			return fmt.Errorf("Gateway %s remains active", kind)
		}
		return nil
	}
}

func TestAccGatewayCoreLifecycle(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	w := accGatewayWorkspace(t)
	client := accGatewayClient(t)
	for _, kind := range []string{"config", "guardrail", "org_guardrail", "service_api_key", "user_api_key", "usage_limit", "rate_limit", "secret_reference", "deployment"} {
		t.Run(kind, func(t *testing.T) {
			name := "tf-gw-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
			addr := "prisma-airs_gateway_" + kind + ".test"
			var rec idRecorder
			ignored := []string{}
			switch kind {
			case "service_api_key", "user_api_key":
				ignored = []string{"key"}
			case "guardrail", "org_guardrail":
				ignored = []string{"check_parameters"}
			case "secret_reference":
				ignored = []string{"auth_config", "allowed_workspaces"}
			case "deployment":
				ignored = []string{"client_auth", "credentials", "deployment_config", "auth_settings"}
			}
			checks := []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)}
			if kind == "config" {
				checks = append(checks, gatewayConfigDeltaCheck{addr: addr, id: &rec.value})
			}
			resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: gatewayDestroy(client, w), Steps: []resource.TestStep{
				{Config: gatewayCoreDiscoveryConfig(kind, name, w), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "name", name), resource.TestCheckResourceAttrSet(addr, "id"), rec.capture("id", addr), gatewayDiscoveryContainsOwned(kind, addr))},
				{ResourceName: addr, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: ignored},
				{Config: gatewayCoreConfig(kind, name, w, true), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: checks}, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "name", name+"-updated"), func(st *terraform.State) error {
					if st.RootModule().Resources[addr].Primary.ID != rec.value {
						return fmt.Errorf("update replaced Gateway resource ID")
					}
					return nil
				})},
				{Config: gatewayCoreConfig(kind, name, w, true), PlanOnly: true},
			}})
		})
	}
}

type gatewayConfigDeltaCheck struct {
	addr string
	id   *string
}

func (c gatewayConfigDeltaCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, change := range req.Plan.ResourceChanges {
		if change.Address != c.addr {
			continue
		}
		before, ok := change.Change.Before.(map[string]any)
		if !ok {
			resp.Error = fmt.Errorf("config before missing")
			return
		}
		after, ok := change.Change.After.(map[string]any)
		if !ok {
			resp.Error = fmt.Errorf("config after missing")
			return
		}
		b, ok := before["config"].(map[string]any)
		if !ok {
			resp.Error = fmt.Errorf("native config before missing")
			return
		}
		a, ok := after["config"].(map[string]any)
		if !ok {
			resp.Error = fmt.Errorf("native config became wholly unknown")
			return
		}
		if after["id"] != *c.id || !reflect.DeepEqual(a["cache"], b["cache"]) || a["provider"] != b["provider"] {
			resp.Error = fmt.Errorf("unchanged routing fields or ID became unknown")
		}
		unknown, ok := change.Change.AfterUnknown.(map[string]any)
		if !ok || unknown["version_id"] != true {
			resp.Error = fmt.Errorf("new config revision must be computed")
		}
		return
	}
	resp.Error = fmt.Errorf("config not present in plan")
}

func gatewayDelete(ctx context.Context, c *gw.Client, kind, id string) error {
	switch kind {
	case "config":
		_, e := c.Configs.Delete(ctx, id)
		return e
	case "guardrail":
		return c.Guardrails.Delete(ctx, id)
	case "org_guardrail":
		return c.OrgGuardrails.Delete(ctx, id)
	case "service_api_key":
		_, e := c.APIKeys.DeleteForKind(ctx, gw.APIKeyService, id)
		return e
	case "user_api_key":
		_, e := c.APIKeys.DeleteForKind(ctx, gw.APIKeyUser, id)
		return e
	case "usage_limit":
		_, e := c.UsageLimits.Delete(ctx, id)
		return e
	case "rate_limit":
		_, e := c.RateLimits.Delete(ctx, id)
		return e
	case "secret_reference":
		_, e := c.SecretReferences.Delete(ctx, id)
		return e
	case "deployment":
		_, e := c.Deployments.Delete(ctx, id)
		return e
	}
	return fmt.Errorf("unsupported external-delete fixture")
}
func TestAccGatewayExternalDeletionAndArchive(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	w := accGatewayWorkspace(t)
	c := accGatewayClient(t)
	for _, kind := range []string{"config", "guardrail", "org_guardrail", "service_api_key", "user_api_key", "usage_limit", "rate_limit", "secret_reference", "deployment"} {
		t.Run(kind, func(t *testing.T) {
			name := "tf-gw-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
			addr := "prisma-airs_gateway_" + kind + ".test"
			var rec idRecorder
			cfg := gatewayCoreConfig(kind, name, w, false)
			resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: gatewayDestroy(c, w), Steps: []resource.TestStep{{Config: cfg, Check: rec.capture("id", addr)}, {Config: cfg, PreConfig: func() {
				ctx, cancel := accCtx()
				defer cancel()
				if e := gatewayDelete(ctx, c, kind, rec.value); e != nil {
					t.Fatal("external deletion failed")
				}
			}, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionCreate)}}, Check: rec.different("id", addr)}, {Config: cfg, PlanOnly: true}}})
		})
	}
}

func TestAccGatewayBindingIsolation(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	w := accGatewayWorkspace(t)
	other := os.Getenv("PANW_AI_GW_TEST_OTHER_WORKSPACE_ID")
	if other == "" || other == w {
		t.Fatal("a distinct existing PANW_AI_GW_TEST_OTHER_WORKSPACE_ID is required")
	}
	c := accGatewayClient(t)
	for _, mcp := range []bool{false, true} {
		t.Run(fmt.Sprintf("mcp=%t", mcp), func(t *testing.T) {
			ctx, cancel := accCtx()
			defer cancel()
			name := "tf-gw-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
			var id string
			if mcp {
				doc, e := s.NewJSONDocument(map[string]any{})
				if e != nil {
					t.Fatal(e)
				}
				r, e := c.MCPIntegrations.Create(ctx, s.CreateMCPIntegration{Name: name, URL: "https://mcp.deepwiki.com/mcp", AuthType: "none", Transport: "http", Configurations: &doc, OrganisationID: stringPointer(os.Getenv("PANW_MGMT_TSG_ID"))})
				if e != nil || r.ID == nil {
					t.Fatal("MCP parent fixture create failed")
				}
				id = *r.ID
			} else {
				r, e := c.Integrations.Create(ctx, s.CreateIntegrationRequest{Name: name, AiProviderID: os.Getenv("PANW_AI_GW_TEST_PROVIDER_ID"), Key: stringPointer("terraform-verification-placeholder"), OrganisationID: stringPointer(os.Getenv("PANW_MGMT_TSG_ID"))})
				if e != nil || r.ID == nil {
					t.Fatal("parent fixture create failed")
				}
				id = *r.ID
			}
			t.Cleanup(func() {
				ctx, cancel := accCtx()
				defer cancel()
				var e error
				if mcp {
					_, e = c.MCPIntegrations.Delete(ctx, id)
				} else {
					_, e = c.Integrations.Delete(ctx, id)
				}
				if e != nil {
					t.Error("parent fixture cleanup failed")
				}
				kind := "integration"
				if mcp {
					kind = "mcp_integration"
				}
				_, e = gatewayRead(ctx, c, kind, id, w)
				if !aisec.IsNotFound(e) {
					t.Error("parent fixture absence unconfirmed")
				}
			})
			body, _ := json.Marshal(map[string]any{"workspaces": []any{map[string]any{"id": other, "enabled": true}}, "create_default_provider": false})
			if mcp {
				var q s.BulkUpdateMCPIntegrationWorkspaces
				if e := json.Unmarshal(body, &q); e != nil {
					t.Fatal(e)
				}
				if _, e := c.MCPIntegrations.SetWorkspaces(ctx, id, q); e != nil {
					t.Fatal("other binding setup failed")
				}
			} else {
				var q s.BulkUpdateWorkspacesRequest
				if e := json.Unmarshal(body, &q); e != nil {
					t.Fatal(e)
				}
				if _, e := c.Integrations.SetWorkspaces(ctx, id, q); e != nil {
					t.Fatal("other binding setup failed")
				}
			}
			kind := "integration_workspace_binding"
			if mcp {
				kind = "mcp_" + kind
			}
			addr := "prisma-airs_gateway_" + kind + ".test"
			config := fmt.Sprintf("resource %q \"test\" {\nintegration_id = %q\nworkspace_id = %q\n}\n", "prisma-airs_gateway_"+kind, id, w)
			otherEnabled := func(*terraform.State) error {
				ctx, cancel := accCtx()
				defer cancel()
				_, e := gatewayRead(ctx, c, kind, id+"/"+other, other)
				if e != nil {
					return fmt.Errorf("unrelated owned fixture binding was removed")
				}
				return nil
			}
			resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: func(st *terraform.State) error {
				if e := gatewayDestroy(c, w)(st); e != nil {
					return e
				}
				return otherEnabled(st)
			}, Steps: []resource.TestStep{{Config: config, Check: otherEnabled}, {ResourceName: addr, ImportState: true, ImportStateVerify: true}, {Config: config, PlanOnly: true}}})
		})
	}
}
func stringPointer(v string) *string { return &v }
func gatewayCoreConfig(kind, name, w string, updated bool) string {
	attempts, value, threshold := 1, 100, 20
	if updated {
		name += "-updated"
		attempts = 0
		value = 0
		threshold = 0
	}
	extra := ""
	switch kind {
	case "config":
		extra = fmt.Sprintf(`workspace_id = %q
config = { provider = "openai", retry = { attempts = %d }, cache = { mode = "simple" } }`, w, attempts)
	case "guardrail", "org_guardrail":
		if kind == "guardrail" {
			extra = fmt.Sprintf("workspace_id = %q\n", w)
		}
		extra += `checks = [{id = "default.isAllLowerCase"}]
check_parameters = { "default.isAllLowerCase" = {} }
actions = { deny = false, async = false, on_success = { feedback = { value = 5, weight = 1, metadata = "" } }, on_fail = { feedback = { value = -5, weight = 1, metadata = "" } } }`
	case "service_api_key", "user_api_key":
		extra = fmt.Sprintf(`workspace_id = %q
scopes = ["completions.write"]
defaults = {allow_config_override = false, metadata = {terraform_test = "owned"}}`, w)
		if kind == "user_api_key" {
			extra += fmt.Sprintf("\nuser_id = %q", os.Getenv("PANW_AI_GW_TEST_USER_ID"))
		}
	case "usage_limit":
		extra = fmt.Sprintf(`workspace_id = %q
type = "tokens"
credit_limit = 100000
alert_threshold = %d
conditions = [{key = "metadata.terraform_test", value = %q}]
group_by = [{key = "metadata.terraform_test"}]`, w, threshold, strings.TrimSuffix(name, "-updated"))
	case "rate_limit":
		extra = fmt.Sprintf(`workspace_id = %q
type = "requests"
unit = "rpm"
value = %d
conditions = [{key = "metadata.terraform_test", value = %q}]
group_by = [{key = "metadata.terraform_test"}]`, w, value, strings.TrimSuffix(name, "-updated"))
	case "secret_reference":
		extra = fmt.Sprintf(`manager_type = "aws_sm"
auth_config = {aws_auth_type = "serviceRole", aws_region = "us-east-1"}
secret_path = "terraform-verification-unbound"
allow_all_workspaces = false
allowed_workspaces = [%q]`, w)
	case "deployment":
		extra = `type = "non_production"
is_default = false
deployment_config = {terraform_test = "owned"}
auth_settings = {is_playground_proxy_allowed = 0}
tags = {terraform_test = "owned"}`
	}
	return fmt.Sprintf("resource %q \"test\" {\nname = %q\n%s\n}\n", "prisma-airs_gateway_"+kind, name, extra)
}

func TestAccGatewayIntegrationGraphs(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	w := accGatewayWorkspace(t)
	client := accGatewayClient(t)
	family := os.Getenv("PANW_AI_GW_TEST_PROVIDER_ID")
	if family == "" {
		t.Fatal("PANW_AI_GW_TEST_PROVIDER_ID required")
	}
	name := "tf-gw-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	steps := []resource.TestStep{{Config: gatewayGraphConfig(name, w, family, false), Check: resource.ComposeAggregateTestCheckFunc(gatewayDiscoveryContainsOwned("integration", "prisma-airs_gateway_integration.test"), gatewayDiscoveryContainsOwned("provider", "prisma-airs_gateway_provider.test"), gatewayDiscoveryContainsOwned("mcp_integration", "prisma-airs_gateway_mcp_integration.test"), gatewayDiscoveryContainsOwned("mcp_server", "prisma-airs_gateway_mcp_server.test"))}, {Config: gatewayGraphConfig(name, w, family, false)}}
	for _, kind := range []string{"integration", "integration_workspace_binding", "provider", "mcp_integration", "mcp_integration_workspace_binding", "mcp_server"} {
		addr := "prisma-airs_gateway_" + kind + ".test"
		ignore := []string{}
		if kind == "integration" {
			ignore = []string{"key"}
		}
		if kind == "mcp_integration" {
			ignore = []string{"configurations"}
		}
		step := resource.TestStep{ResourceName: addr, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: ignore}
		if kind == "provider" || kind == "mcp_server" {
			step.ImportStateIdFunc = func(st *terraform.State) (string, error) {
				r := st.RootModule().Resources[addr]
				return w + "/" + r.Primary.ID, nil
			}
		}
		steps = append(steps, step)
	}
	steps = append(steps, resource.TestStep{Config: gatewayGraphConfig(name, w, family, true)}, resource.TestStep{Config: gatewayGraphConfig(name, w, family, true), PlanOnly: true})
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: gatewayDestroy(client, w), Steps: steps})
}
func gatewayGraphConfig(name, w, family string, updated bool) string {
	desc := "initial"
	if updated {
		desc = "updated"
	}
	return fmt.Sprintf(`
resource "prisma-airs_gateway_integration" "test" {
 name = %[1]q
 ai_provider_id = %[3]q
 key = "terraform-verification-placeholder"
 description = %[4]q
 secret_mappings = []
 pricing_adjustments = {multiplier = {request_token = 1, response_token = 1}}
}
resource "prisma-airs_gateway_integration_workspace_binding" "test" {
 integration_id = prisma-airs_gateway_integration.test.id
 workspace_id = %[2]q
}
resource "prisma-airs_gateway_provider" "test" {
 name = %[1]q
 integration_id = prisma-airs_gateway_integration.test.id
 workspace_id = %[2]q
 note = %[4]q
 usage_limits = {type = "tokens", credit_limit = 100000, alert_threshold = 20}
 depends_on = [prisma-airs_gateway_integration_workspace_binding.test]
}
resource "prisma-airs_gateway_mcp_integration" "test" {
 name = %[1]q
 url = "https://mcp.deepwiki.com/mcp"
 auth_type = "none"
 transport = "http"
 configurations = {}
 secret_mappings = []
 description = %[4]q
}
resource "prisma-airs_gateway_mcp_integration_workspace_binding" "test" {
 integration_id = prisma-airs_gateway_mcp_integration.test.id
 workspace_id = %[2]q
}
resource "prisma-airs_gateway_mcp_server" "test" {
 name = %[1]q
 mcp_integration_id = prisma-airs_gateway_mcp_integration.test.id
 workspace_id = %[2]q
 description = %[4]q
 depends_on = [prisma-airs_gateway_mcp_integration_workspace_binding.test]
}
`, name, w, family, desc) + fmt.Sprintf(`
 data "prisma-airs_gateway_integrations" "owned" { depends_on = [prisma-airs_gateway_integration.test] }
 data "prisma-airs_gateway_providers" "owned" {
  workspace_id = %q
  depends_on = [prisma-airs_gateway_provider.test]
 }
 data "prisma-airs_gateway_mcp_integrations" "owned" { depends_on = [prisma-airs_gateway_mcp_integration.test] }
 data "prisma-airs_gateway_mcp_servers" "owned" {
  workspace_id = %q
  depends_on = [prisma-airs_gateway_mcp_server.test]
 }
 `, w, w)
}

func TestAccGatewayDiscovery(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	w := accGatewayWorkspace(t)
	config := ""
	for _, kind := range []string{"configs", "guardrails", "org_guardrails", "integrations", "providers", "mcp_integrations", "mcp_servers", "service_api_keys", "user_api_keys", "usage_limits", "rate_limits", "secret_references", "deployments"} {
		attrs := ""
		switch kind {
		case "configs", "guardrails", "providers", "mcp_servers", "service_api_keys", "user_api_keys", "usage_limits", "rate_limits":
			attrs = fmt.Sprintf("workspace_id = %q", w)
		}
		config += fmt.Sprintf("data %q \"test\" {\n%s\n}\n", "prisma-airs_gateway_"+kind, attrs)
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{{Config: config}, {Config: config, PlanOnly: true}}})
}

// Exercise the authored example, retaining its routing document and graph.
// Only its unpublished version constraint, variable defaults and names are
// substituted for the in-process test provider and disposable fixtures.
func TestAccGatewayPublishedExample(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "cdot65")
	w := accGatewayWorkspace(t)
	family := os.Getenv("PANW_AI_GW_TEST_PROVIDER_ID")
	if family == "" {
		t.Fatal("PANW_AI_GW_TEST_PROVIDER_ID required")
	}
	source, err := os.ReadFile("../../examples/gateway/main.tf")
	if err != nil {
		t.Fatal(err)
	}
	name := "tf-gw-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	config := strings.ReplaceAll(string(source), `, version = "~> 0.9.0"`, "")
	config = strings.ReplaceAll(config, `variable "workspace_id" { type = string }`, fmt.Sprintf(`variable "workspace_id" {
 type = string
 default = %q
}`, w))
	config = strings.ReplaceAll(config, `variable "ai_provider_id" { type = string }`, fmt.Sprintf(`variable "ai_provider_id" {
 type = string
 default = %q
}`, family))
	config = strings.ReplaceAll(config, `variable "model" { type = string }`, `variable "model" {
 type = string
 default = "gpt-4o"
}`)
	config = strings.ReplaceAll(config, "  sensitive = true", "  sensitive = true\n  default = \"terraform-verification-placeholder\"")
	for _, old := range []string{"Example - Gateway - Development", "Example - Provider - Development", "Example - Routing - Development"} {
		config = strings.ReplaceAll(config, old, name)
	}
	providerUpdated := strings.Replace(config, `resource "prisma-airs_gateway_provider" "application" {`, `resource "prisma-airs_gateway_provider" "application" {
 note = "updated without routing revision"`, 1)
	leafUpdated := strings.Replace(providerUpdated, "retry = { attempts = 1 }", "retry = { attempts = 2 }", 1)
	addr := "prisma-airs_gateway_config.application"
	var revision idRecorder
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, CheckDestroy: gatewayDestroy(accGatewayClient(t), w), Steps: []resource.TestStep{
		{Config: config, Check: revision.capture("version_id", addr)},
		{Config: config}, // Bindings update parent metadata; refresh before the empty plan.
		{Config: config, PlanOnly: true},
		{Config: providerUpdated, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionNoop)}}, Check: func(st *terraform.State) error {
			if st.RootModule().Resources[addr].Primary.Attributes["version_id"] != revision.value {
				return fmt.Errorf("provider note change created a config revision")
			}
			return nil
		}},
		{Config: leafUpdated, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)}}, Check: func(st *terraform.State) error {
			if st.RootModule().Resources[addr].Primary.Attributes["version_id"] == revision.value {
				return fmt.Errorf("leaf edit did not change revision")
			}
			return nil
		}},
		{Config: leafUpdated, PlanOnly: true},
	}})
}

func TestGatewayDestroySkipsMetadataDataSources(t *testing.T) {
	state := &terraform.State{Modules: []*terraform.ModuleState{{Path: []string{"root"}, Resources: map[string]*terraform.ResourceState{"prisma-airs_gateway_configs.workspace": {Type: "prisma-airs_gateway_configs", Primary: &terraform.InstanceState{ID: "discovery", Attributes: map[string]string{"total_count": "0"}}}}}}}
	if err := gatewayDestroy(nil, "workspace")(state); err != nil {
		t.Fatal(err)
	}
}

func gatewayCoreDiscoveryConfig(kind, name, w string) string {
	plural := kind + "s"
	attrs := ""
	switch kind {
	case "config", "guardrail", "service_api_key", "user_api_key", "usage_limit", "rate_limit":
		attrs = fmt.Sprintf("workspace_id = %q", w)
	}
	return gatewayCoreConfig(kind, name, w, false) + fmt.Sprintf(`
 data "prisma-airs_gateway_%s" "owned" {
  %s
  depends_on = [prisma-airs_gateway_%s.test]
 }
 `, plural, attrs, kind)
}
func gatewayDiscoveryContainsOwned(kind, addr string) resource.TestCheckFunc {
	return func(st *terraform.State) error {
		data := st.RootModule().Resources["data.prisma-airs_gateway_"+kind+"s.owned"]
		if data == nil {
			return fmt.Errorf("owned discovery source missing")
		}
		n, err := strconv.Atoi(data.Primary.Attributes["items.#"])
		if err != nil {
			return err
		}
		id := st.RootModule().Resources[addr].Primary.ID
		for i := 0; i < n; i++ {
			if data.Primary.Attributes[fmt.Sprintf("items.%d.id", i)] == id {
				return nil
			}
		}
		return fmt.Errorf("first Gateway discovery page did not include the owned %s", kind)
	}
}
