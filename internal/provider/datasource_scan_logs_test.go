package provider_test

import (
	"testing"
)

func TestAccScanLogsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: the Scan Logs API returns 403 "Access denied" for the current
	// service account.
	t.Skip("Scan Logs API returns 403 with current service account permissions")
}
