package supplychain

import (
	"os"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/cdot65/prisma-airs-go/aisec/agentguard"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Clients holds the two SDK services owned by the Supply Chain product.
type Clients struct {
	Models      *modelsecurity.Client
	Skills      *agentguard.Client
	TenantID    string
	SkillsError error
}

func Definition() product.Definition {
	return product.Definition{
		ID: "supply_chain", Label: "AI Supply Chain Security", Implemented: true,
		Endpoints: []product.Endpoint{
			{Name: "data_endpoint", Environment: "PANW_MODEL_SEC_DATA_ENDPOINT", Description: "Supply Chain Security model-management data endpoint override."},
			{Name: "mgmt_endpoint", Environment: "PANW_MODEL_SEC_MGMT_ENDPOINT", Description: "Supply Chain Security model-management endpoint override."},
			{Name: "skill_scanning_data_endpoint", Environment: "PANW_SKILL_SCANNING_DATA_ENDPOINT", Description: "Skill Scanning data base URL, including its product prefix. Both Skill Scanning endpoints are required when using its resources."},
			{Name: "skill_scanning_mgmt_endpoint", Environment: "PANW_SKILL_SCANNING_MGMT_ENDPOINT", Description: "Skill Scanning management base URL, including its product prefix."},
		},
		Resources: []product.Resource{
			{New: NewModelSecurityGroupResource, Guide: "resources/model-security-group"},
			{New: NewSkillScanningInstanceResource, Guide: "resources/skill-scanning-instance"},
			{New: NewSkillScanningRuleResource, Guide: "resources/skill-scanning-rule"},
			{New: NewSkillScanningOverrideResource, Guide: "resources/skill-scanning-override"},
		},
		DataSources: skillDataSources(),
		Configure: func(c product.Credentials, endpoints map[string]string) (any, error) {
			models, err := modelsecurity.NewClient(modelsecurity.Opts{ClientID: c.ClientID, ClientSecret: c.ClientSecret, TsgID: c.TsgID,
				TokenEndpoint: c.TokenEndpoint, DataEndpoint: endpoints["data_endpoint"], MgmtEndpoint: endpoints["mgmt_endpoint"]})
			if err != nil {
				return nil, err
			}
			clients := &Clients{Models: models, TenantID: c.TsgID}
			data, mgmt := endpoints["skill_scanning_data_endpoint"], endpoints["skill_scanning_mgmt_endpoint"]
			// The SDK's established endpoint variables remain useful to SDK consumers.
			if data == "" {
				data = os.Getenv("PANW_AGENT_GUARD_DATA_ENDPOINT")
			}
			if mgmt == "" {
				mgmt = os.Getenv("PANW_AGENT_GUARD_MGMT_ENDPOINT")
			}
			if data != "" || mgmt != "" {
				tokenEndpoint := c.TokenEndpoint
				if tokenEndpoint == "" {
					tokenEndpoint = aisec.DefaultTokenEndpoint
				}
				clients.Skills, err = agentguard.NewClient(agentguard.Opts{ClientID: c.ClientID, ClientSecret: c.ClientSecret, TsgID: c.TsgID, TokenEndpoint: tokenEndpoint, DataEndpoint: data, MgmtEndpoint: mgmt})
				if err != nil {
					clients.SkillsError = err
				}
			}
			return clients, nil
		},
	}
}

func getModelSecClient(data any) (*modelsecurity.Client, diag.Diagnostics) {
	clients, diagnostics := product.Client[*Clients](data, "supply_chain", "AI Supply Chain Security")
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	return clients.Models, diagnostics
}

func getSkillClient(data any) (*Clients, diag.Diagnostics) {
	clients, diagnostics := product.Client[*Clients](data, "supply_chain", "AI Supply Chain Security")
	if !diagnostics.HasError() && clients.SkillsError != nil {
		diagnostics.AddError("Skill Scanning endpoints are invalid", "Configure both Skill Scanning base URLs. "+skillError(clients.SkillsError))
	} else if !diagnostics.HasError() && clients.Skills == nil {
		diagnostics.AddError("Skill Scanning is not configured", "Set supply_chain.skill_scanning_data_endpoint and supply_chain.skill_scanning_mgmt_endpoint, or PANW_SKILL_SCANNING_DATA_ENDPOINT and PANW_SKILL_SCANNING_MGMT_ENDPOINT. Both base URLs are required; shared OAuth credentials are used.")
	}
	return clients, diagnostics
}
