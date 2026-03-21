package provider_test

import (
	"os"
	"testing"

	"github.com/cdot65/prisma-airs-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"prisma-airs": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("PANW_MGMT_CLIENT_ID") == "" {
		t.Skip("PANW_MGMT_CLIENT_ID not set")
	}
	if os.Getenv("PANW_MGMT_CLIENT_SECRET") == "" {
		t.Skip("PANW_MGMT_CLIENT_SECRET not set")
	}
	if os.Getenv("PANW_MGMT_TSG_ID") == "" {
		t.Skip("PANW_MGMT_TSG_ID not set")
	}
}
