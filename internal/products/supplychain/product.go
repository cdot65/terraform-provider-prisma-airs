package supplychain

import (
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func Definition() product.Definition {
	return product.Definition{
		ID: "supply_chain", Label: "AI Supply Chain Security", Implemented: true,
		Endpoints: []product.Endpoint{
			{Name: "data_endpoint", Environment: "PANW_MODEL_SEC_DATA_ENDPOINT", Description: "Supply Chain Security model-management data endpoint override."},
			{Name: "mgmt_endpoint", Environment: "PANW_MODEL_SEC_MGMT_ENDPOINT", Description: "Supply Chain Security model-management endpoint override."},
		},
		Resources:   []product.Resource{{New: NewModelSecurityGroupResource, Guide: "resources/model-security-group"}},
		DataSources: []product.DataSource{{New: NewModelSecurityRulesDataSource, Guide: "data-sources/model-security-rules"}},
		Configure: func(c product.Credentials, endpoints map[string]string) (any, error) {
			return modelsecurity.NewClient(modelsecurity.Opts{ClientID: c.ClientID, ClientSecret: c.ClientSecret, TsgID: c.TsgID,
				TokenEndpoint: c.TokenEndpoint, DataEndpoint: endpoints["data_endpoint"], MgmtEndpoint: endpoints["mgmt_endpoint"]})
		},
	}
}

func getModelSecClient(data any) (*modelsecurity.Client, diag.Diagnostics) {
	return product.Client[*modelsecurity.Client](data, "supply_chain", "AI Supply Chain Security")
}
