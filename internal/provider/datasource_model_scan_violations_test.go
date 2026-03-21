package provider_test

import (
	"testing"
)

func TestAccModelScanViolationsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: requires a completed scan UUID to fetch violations.
	t.Skip("Model scan violations require a completed scan UUID")
}
