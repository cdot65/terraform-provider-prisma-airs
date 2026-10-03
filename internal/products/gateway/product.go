package gateway

import (
	"strings"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Definition() product.Definition {
	d := product.Definition{ID: "gateway", Label: "AI Gateway", Implemented: true,
		Endpoints: []product.Endpoint{{Name: "data_endpoint", Environment: "PANW_AI_GW_DATA_ENDPOINT", Description: "Gateway data-plane management endpoint override."}, {Name: "admin_endpoint", Environment: "PANW_AI_GW_ADMIN_ENDPOINT", Description: "Gateway admin-plane management endpoint override."}, {Name: "iam_endpoint", Environment: "PANW_IAM_ENDPOINT", Description: "SCM IAM endpoint override for workspace scope orchestration."}},
		Configure: func(c product.Credentials, endpoints map[string]string) (any, error) {
			sdk, e := gw.NewClient(gw.Opts{ClientID: c.ClientID, ClientSecret: c.ClientSecret, TsgID: c.TsgID, TokenEndpoint: c.TokenEndpoint, DataEndpoint: endpoints["data_endpoint"], AdminEndpoint: endpoints["admin_endpoint"], IAMEndpoint: endpoints["iam_endpoint"]})
			if e != nil {
				return nil, e
			}
			return &client{sdk: sdk, organisation: c.TsgID}, nil
		},
	}
	for _, entry := range definitions() {
		d.Resources = append(d.Resources, product.Resource{New: func() resource.Resource { return &gatewayResource{definition: entry} }, Guide: "resources/gateway-" + strings.ReplaceAll(entry.name, "_", "-")})
		if entry.list != nil {
			d.DataSources = append(d.DataSources, product.DataSource{New: func() datasource.DataSource { return &gatewayDataSource{definition: entry} }, Guide: "data-sources/gateway-" + strings.ReplaceAll(plural(entry.name), "_", "-")})
		}
	}
	d.Resources = append(d.Resources, product.Resource{New: func() resource.Resource { return &workspaceResource{} }, Guide: "resources/gateway-workspace"})
	d.DataSources = append(d.DataSources, product.DataSource{New: func() datasource.DataSource { return &workspaceDataSource{} }, Guide: "data-sources/gateway-workspace"}, product.DataSource{New: func() datasource.DataSource { return &workspaceDataSource{list: true} }, Guide: "data-sources/gateway-workspaces"})
	return d
}
