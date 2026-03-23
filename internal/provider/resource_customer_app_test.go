package provider_test

import (
	"testing"
)

func TestAccCustomerAppResource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: the Customer Apps API returns 403 "Access denied" for the current
	// service account. Both POST /v1/mgmt/customerapp and GET /v1/mgmt/customerapps
	// are forbidden.
	t.Skip("Customer Apps API returns 403 with current service account permissions")
}
