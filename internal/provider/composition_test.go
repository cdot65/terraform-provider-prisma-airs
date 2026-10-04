package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec/agentguard"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/cdot65/prisma-airs-provider/internal/products"
	"github.com/cdot65/prisma-airs-provider/internal/products/supplychain"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func configureFixture(t *testing.T, values map[string]string, endpoints map[string]map[string]string) provider.ConfigureResponse {
	t.Helper()
	ctx := context.Background()
	p := New("test")()
	var schema provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schema)
	blocks := map[string]types.Object{}
	for _, definition := range products.All() {
		if !definition.Implemented {
			continue
		}
		attrs := map[string]attr.Type{}
		vals := map[string]attr.Value{}
		for _, endpoint := range definition.Endpoints {
			attrs[endpoint.Name] = types.StringType
			vals[endpoint.Name] = types.StringNull()
			if value, ok := endpoints[definition.ID][endpoint.Name]; ok {
				vals[endpoint.Name] = types.StringValue(value)
			}
		}
		blocks[definition.ID] = types.ObjectNull(attrs)
		if _, ok := endpoints[definition.ID]; ok {
			blocks[definition.ID] = types.ObjectValueMust(attrs, vals)
		}
	}
	stringValue := func(name string) types.String {
		if value, ok := values[name]; ok {
			return types.StringValue(value)
		}
		return types.StringNull()
	}
	model := PrismaAIRSProviderModel{ClientID: stringValue("client_id"), ClientSecret: stringValue("client_secret"), TsgID: stringValue("tsg_id"), TokenEndpoint: stringValue("token_endpoint"), Runtime: blocks["runtime"], RedTeam: blocks["red_team"], SupplyChain: blocks["supply_chain"], Gateway: blocks["gateway"]}
	state := tfsdk.State{Schema: schema.Schema}
	if diagnostics := state.Set(ctx, model); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	var response provider.ConfigureResponse
	p.Configure(ctx, provider.ConfigureRequest{Config: tfsdk.Config{Schema: schema.Schema, Raw: state.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	return response
}

func TestProductCatalogMatchesProtocolSchema(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(New("test")())()
	response, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, diagnostic := range response.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatal(diagnostic.Detail)
		}
	}
	wantLabels := map[string]string{"runtime": "AI Runtime Security", "red_team": "AI Red Teaming", "gateway": "AI Gateway", "supply_chain": "AI Supply Chain Security"}
	resources := map[string]bool{}
	sources := map[string]bool{}
	for _, definition := range products.All() {
		if wantLabels[definition.ID] != definition.Label {
			t.Errorf("unexpected product %s", definition.ID)
		}
		delete(wantLabels, definition.ID)
		metadata := definition.Metadata(ctx, "prisma-airs")
		if !definition.Implemented && (len(definition.Resources) > 0 || len(definition.DataSources) > 0 || len(definition.Endpoints) > 0 || definition.Configure != nil) {
			t.Error("unimplemented product accepts functionality")
		}
		for _, entry := range metadata.Resources {
			if resources[entry.Name] || response.ResourceSchemas[entry.Name] == nil || !strings.HasPrefix(entry.Name, "prisma-airs_"+definition.ID+"_") {
				t.Errorf("invalid ownership: %s", entry.Name)
			}
			resources[entry.Name] = true
		}
		for _, entry := range metadata.DataSources {
			if sources[entry.Name] || response.DataSourceSchemas[entry.Name] == nil || !strings.HasPrefix(entry.Name, "prisma-airs_"+definition.ID+"_") {
				t.Errorf("invalid ownership: %s", entry.Name)
			}
			sources[entry.Name] = true
		}
	}
	if len(wantLabels) != 0 || len(resources) != 27 || len(sources) != 30 || len(response.ResourceSchemas) != len(resources) || len(response.DataSourceSchemas) != len(sources) {
		t.Fatal("catalog has missing or unclassified types")
	}
	if len(response.Provider.Block.Attributes) != 4 || len(response.Provider.Block.BlockTypes) != 4 {
		t.Fatal("expected shared credentials and four implemented product blocks")
	}
	for _, attribute := range response.Provider.Block.Attributes {
		if attribute.Name == "client_secret" && !attribute.Sensitive {
			t.Error("client secret must be sensitive")
		}
	}
	for _, block := range response.Provider.Block.BlockTypes {
		if block.TypeName == "gateway" && len(block.Block.Attributes) != 3 {
			t.Error("Gateway requires data, admin and IAM endpoint settings")
		}
	}
}

func TestProductEndpointAndCredentialRouting(t *testing.T) {
	for _, name := range []string{"environment and omitted blocks", "explicit nested overrides", "partial nested overrides"} {
		explicit := name == "explicit nested overrides"
		t.Run(name, func(t *testing.T) {
			var lock sync.Mutex
			requests := map[string]int{}
			handler := func(label string) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					lock.Lock()
					requests[label]++
					lock.Unlock()
					w.Header().Set("Content-Type", "application/json")
					if label == "token" {
						if err := r.ParseForm(); err != nil {
							t.Error(err)
						}
						clientID, clientSecret, _ := r.BasicAuth()
						if clientID != "fixture-client" || clientSecret != "fixture-secret" || r.Form.Get("scope") != "tsg_id:fixture-tsg" {
							t.Error("shared credentials not forwarded")
						}
						_, _ = w.Write([]byte(`{"access_token":"fixture-token","token_type":"Bearer","expires_in":3600}`))
						return
					}
					if strings.HasPrefix(label, "gateway-") && r.Header.Get("x-tsg-id") != "fixture-tsg" {
						t.Error("Gateway shared TSG header missing")
					}
					if r.Header.Get("Authorization") != "Bearer fixture-token" {
						t.Error("token not used")
					}
					_, _ = w.Write([]byte(`{"items":[],"data":[]}`))
				}
			}
			servers := map[string]*httptest.Server{}
			for _, label := range []string{"token", "runtime", "red-data", "red-mgmt", "supply-data", "supply-mgmt", "gateway-data", "gateway-admin"} {
				servers[label] = httptest.NewServer(handler(label))
				t.Cleanup(servers[label].Close)
			}
			values := map[string]string{"client_id": "fixture-client", "client_secret": "fixture-secret", "tsg_id": "fixture-tsg", "token_endpoint": servers["token"].URL}
			envs := map[string]string{"PANW_MGMT_CLIENT_ID": values["client_id"], "PANW_MGMT_CLIENT_SECRET": values["client_secret"], "PANW_MGMT_TSG_ID": values["tsg_id"], "PANW_MGMT_TOKEN_ENDPOINT": values["token_endpoint"], "PANW_MGMT_ENDPOINT": servers["runtime"].URL, "PANW_RED_TEAM_DATA_ENDPOINT": servers["red-data"].URL, "PANW_RED_TEAM_MGMT_ENDPOINT": servers["red-mgmt"].URL, "PANW_MODEL_SEC_DATA_ENDPOINT": servers["supply-data"].URL, "PANW_MODEL_SEC_MGMT_ENDPOINT": servers["supply-mgmt"].URL, "PANW_AI_GW_DATA_ENDPOINT": servers["gateway-data"].URL, "PANW_AI_GW_ADMIN_ENDPOINT": servers["gateway-admin"].URL, "PANW_SKILL_SCANNING_DATA_ENDPOINT": servers["supply-data"].URL, "PANW_SKILL_SCANNING_MGMT_ENDPOINT": servers["supply-mgmt"].URL}
			for key, value := range envs {
				if explicit {
					t.Setenv(key, "wrong-environment-value")
				} else {
					t.Setenv(key, value)
				}
			}
			var settings map[string]map[string]string
			if explicit {
				settings = map[string]map[string]string{"runtime": {"mgmt_endpoint": servers["runtime"].URL}, "red_team": {"data_endpoint": servers["red-data"].URL, "mgmt_endpoint": servers["red-mgmt"].URL}, "supply_chain": {"data_endpoint": servers["supply-data"].URL, "mgmt_endpoint": servers["supply-mgmt"].URL, "skill_scanning_data_endpoint": servers["supply-data"].URL, "skill_scanning_mgmt_endpoint": servers["supply-mgmt"].URL}, "gateway": {"data_endpoint": servers["gateway-data"].URL, "admin_endpoint": servers["gateway-admin"].URL}}
			} else {
				values = nil
				if name == "partial nested overrides" {
					t.Setenv("PANW_RED_TEAM_MGMT_ENDPOINT", "wrong-environment-value")
					t.Setenv("PANW_MODEL_SEC_MGMT_ENDPOINT", "wrong-environment-value")
					t.Setenv("PANW_AI_GW_ADMIN_ENDPOINT", "wrong-environment-value")
					settings = map[string]map[string]string{"runtime": {}, "red_team": {"mgmt_endpoint": servers["red-mgmt"].URL}, "supply_chain": {"mgmt_endpoint": servers["supply-mgmt"].URL}, "gateway": {"admin_endpoint": servers["gateway-admin"].URL}}
				}
			}
			configured := configureFixture(t, values, settings)
			if len(requests) != 0 {
				t.Fatal("Configure must not perform network operations")
			}
			data := configured.ResourceData.(product.Data)
			ctx := context.Background()
			runtime := data["runtime"].(*airsruntime.Client)
			red := data["red_team"].(*redteam.Client)
			supply := data["supply_chain"].(*supplychain.Clients).Models
			if _, err := runtime.DlpProfiles.List(ctx, airsruntime.ListOpts{}); err != nil {
				t.Fatal(err)
			}
			if _, err := red.Targets.List(ctx, redteam.TargetListOpts{}); err != nil {
				t.Fatal(err)
			}
			if _, err := red.Scans.List(ctx, redteam.ScanListOpts{}); err != nil {
				t.Fatal(err)
			}
			if _, err := supply.SecurityGroups.List(ctx, modelsecurity.GroupListOpts{}); err != nil {
				t.Fatal(err)
			}
			if _, err := supply.Scans.List(ctx, modelsecurity.ScanListOpts{}); err != nil {
				t.Fatal(err)
			}

			skills := data["supply_chain"].(*supplychain.Clients).Skills
			if _, err := skills.Rules.List(ctx, agentguard.ListOpts{}); err != nil {
				t.Fatal(err)
			}
			if _, err := skills.Scans.List(ctx, agentguard.ScanListOpts{}); err != nil {
				t.Fatal(err)
			}
			for _, definition := range products.All() {
				if definition.ID != "gateway" {
					continue
				}
				for _, index := range []int{0, 2} {
					source := definition.DataSources[index].New()
					var sr datasource.SchemaResponse
					source.Schema(ctx, datasource.SchemaRequest{}, &sr)
					attrs := map[string]attr.Value{}
					ts := map[string]attr.Type{}
					for key, a := range sr.Schema.Attributes {
						ts[key] = a.GetType()
						v, e := ts[key].ValueFromTerraform(ctx, tftypes.NewValue(ts[key].TerraformType(ctx), nil))
						if e != nil {
							t.Fatal(e)
						}
						attrs[key] = v
					}
					if _, ok := attrs["workspace_id"]; ok {
						attrs["workspace_id"] = types.StringValue("fixture-workspace")
					}
					model := types.ObjectValueMust(ts, attrs)
					state := tfsdk.State{Schema: sr.Schema}
					if d := state.Set(ctx, model); d.HasError() {
						t.Fatal(d)
					}
					var cr datasource.ConfigureResponse
					source.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{ProviderData: data}, &cr)
					if cr.Diagnostics.HasError() {
						t.Fatal(cr.Diagnostics)
					}
					rr := datasource.ReadResponse{State: tfsdk.State{Schema: sr.Schema}}
					source.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: state.Raw}}, &rr)
					if rr.Diagnostics.HasError() {
						t.Fatal(rr.Diagnostics)
					}
				}
			}
			for _, label := range []string{"token", "runtime", "red-data", "red-mgmt", "supply-data", "supply-mgmt", "gateway-data", "gateway-admin"} {
				if requests[label] == 0 {
					t.Errorf("endpoint %s not reached", label)
				}
			}
			// Exercise every framework Configure callback, including a deliberately
			// wrong product slot: a generic client bag must not silently cross-wire SDKs.
			for _, definition := range products.All() {
				for _, entry := range definition.Resources {
					r := entry.New().(resource.ResourceWithConfigure)
					var ok resource.ConfigureResponse
					r.Configure(ctx, resource.ConfigureRequest{ProviderData: configured.ResourceData}, &ok)
					if ok.Diagnostics.HasError() {
						t.Fatal(ok.Diagnostics)
					}
					wrong := product.Data{definition.ID: "wrong SDK type"}
					var bad resource.ConfigureResponse
					r.Configure(ctx, resource.ConfigureRequest{ProviderData: wrong}, &bad)
					if !bad.Diagnostics.HasError() {
						t.Error("wrong SDK type accepted")
					}
				}
				for _, entry := range definition.DataSources {
					d := entry.New().(datasource.DataSourceWithConfigure)
					var ok datasource.ConfigureResponse
					d.Configure(ctx, datasource.ConfigureRequest{ProviderData: configured.DataSourceData}, &ok)
					if ok.Diagnostics.HasError() {
						t.Fatal(ok.Diagnostics)
					}
					var invalid datasource.ConfigureResponse
					d.Configure(ctx, datasource.ConfigureRequest{ProviderData: product.Data{definition.ID: "wrong SDK type"}}, &invalid)
					if !invalid.Diagnostics.HasError() {
						t.Error("wrong SDK type accepted by data source")
					}

				}
			}
		})
	}
}

func TestMissingCredentialsAreReportedByUsedProduct(t *testing.T) {
	for _, key := range []string{"PANW_MGMT_CLIENT_ID", "PANW_MGMT_CLIENT_SECRET", "PANW_MGMT_TSG_ID"} {
		t.Setenv(key, "")
	}
	configured := configureFixture(t, nil, nil)
	if len(configured.ResourceData.(product.Data)) != 0 {
		t.Fatal("incomplete credentials constructed clients")
	}
	for _, definition := range products.All() {
		for _, entry := range definition.Resources {
			var response resource.ConfigureResponse
			entry.New().(resource.ResourceWithConfigure).Configure(context.Background(), resource.ConfigureRequest{ProviderData: configured.ResourceData}, &response)
			if !response.Diagnostics.HasError() || !strings.Contains(response.Diagnostics.Errors()[0].Summary(), definition.Label) {
				t.Errorf("missing product diagnostic: %s", definition.ID)
			}
		}
	}
}

// Validation calls Configure before a real provider configuration is available.
// Exercise the public framework callbacks rather than mirrored client helpers.
func TestConfigureBeforeProviderConfiguration(t *testing.T) {
	ctx := context.Background()
	for _, definition := range products.All() {
		for _, entry := range definition.Resources {
			r := entry.New().(resource.ResourceWithConfigure)
			var pending resource.ConfigureResponse
			r.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &pending)
			if pending.Diagnostics.HasError() {
				t.Fatal(pending.Diagnostics)
			}
			var invalid resource.ConfigureResponse
			r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid provider data"}, &invalid)
			if !invalid.Diagnostics.HasError() {
				t.Error("untyped provider data accepted")
			}
		}
		for _, entry := range definition.DataSources {
			d := entry.New().(datasource.DataSourceWithConfigure)
			var pending datasource.ConfigureResponse
			d.Configure(ctx, datasource.ConfigureRequest{ProviderData: nil}, &pending)
			if pending.Diagnostics.HasError() {
				t.Fatal(pending.Diagnostics)
			}
			var invalid datasource.ConfigureResponse
			d.Configure(ctx, datasource.ConfigureRequest{ProviderData: "invalid provider data"}, &invalid)
			if !invalid.Diagnostics.HasError() {
				t.Error("untyped provider data accepted")
			}
		}
	}
}

func TestImplementedProductsHaveConfigurationModels(t *testing.T) {
	fields := map[string]bool{}
	model := reflect.TypeOf(PrismaAIRSProviderModel{})
	for i := 0; i < model.NumField(); i++ {
		fields[model.Field(i).Tag.Get("tfsdk")] = true
	}
	for _, definition := range products.All() {
		if definition.Implemented && !fields[definition.ID] {
			t.Errorf("missing configuration model for %s", definition.ID)
		}
	}
}
