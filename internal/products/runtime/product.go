package runtime

import (
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func Definition() product.Definition {
	return product.Definition{
		ID: "runtime", Label: "AI Runtime Security", Implemented: true,
		Endpoints: []product.Endpoint{{Name: "mgmt_endpoint", Environment: "PANW_MGMT_ENDPOINT", Description: "Runtime Security management endpoint override."}},
		Resources: []product.Resource{
			{New: NewSecurityProfileResource, Guide: "resources/security-profile"},
			{New: NewCustomTopicResource, Guide: "resources/custom-topic"},
			{New: NewApiKeyResource, Guide: "resources/api-key"},
			{New: NewCustomerAppResource, Guide: "resources/customer-app"},
		},
		DataSources: []product.DataSource{
			{New: NewDlpProfilesDataSource, Guide: "data-sources/dlp-profiles"},
			{New: NewDeploymentProfilesDataSource, Guide: "data-sources/deployment-profiles"},
		},
		Configure: func(c product.Credentials, endpoints map[string]string) (any, error) {
			return airsruntime.NewClient(airsruntime.Opts{ClientID: c.ClientID, ClientSecret: c.ClientSecret, TsgID: c.TsgID,
				TokenEndpoint: c.TokenEndpoint, APIEndpoint: endpoints["mgmt_endpoint"]})
		},
	}
}

func getMgmtClient(data any) (*airsruntime.Client, diag.Diagnostics) {
	return product.Client[*airsruntime.Client](data, "runtime", "AI Runtime Security")
}
