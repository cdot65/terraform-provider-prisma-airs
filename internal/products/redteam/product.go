package redteam

import (
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func Definition() product.Definition {
	return product.Definition{
		ID: "red_team", Label: "AI Red Teaming", Implemented: true,
		Endpoints: []product.Endpoint{
			{Name: "data_endpoint", Environment: "PANW_RED_TEAM_DATA_ENDPOINT", Description: "Red Teaming data endpoint override."},
			{Name: "mgmt_endpoint", Environment: "PANW_RED_TEAM_MGMT_ENDPOINT", Description: "Red Teaming management endpoint override."},
		},
		Resources: []product.Resource{
			{New: NewRedTeamTargetResource, Guide: "resources/red-team-target"},
			{New: NewRedTeamCustomPromptSetResource, Guide: "resources/red-team-custom-prompt-set"},
		},
		Configure: func(c product.Credentials, endpoints map[string]string) (any, error) {
			return redteam.NewClient(redteam.Opts{ClientID: c.ClientID, ClientSecret: c.ClientSecret, TsgID: c.TsgID,
				TokenEndpoint: c.TokenEndpoint, DataEndpoint: endpoints["data_endpoint"], MgmtEndpoint: endpoints["mgmt_endpoint"]})
		},
	}
}

func getRedTeamClient(data any) (*redteam.Client, diag.Diagnostics) {
	return product.Client[*redteam.Client](data, "red_team", "AI Red Teaming")
}
