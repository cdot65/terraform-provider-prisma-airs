package provider_test

import (
	"testing"
)

func TestAccApiKeyResource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: API key creation requires auth_code and customer_app_name fields.
	// The Customer Apps API returns 403 for the current service account,
	// so we cannot create customer apps needed for API key creation.
	t.Skip("API key creation requires customer apps which return 403 with current service account")
}
