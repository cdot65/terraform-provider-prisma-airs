package provider_test

import (
	"testing"
)

func TestAccModelScanEvaluationsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: requires a completed scan UUID to fetch evaluations.
	t.Skip("Model scan evaluations require a completed scan UUID")
}
