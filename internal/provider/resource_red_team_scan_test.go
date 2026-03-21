package provider_test

import (
	"testing"
)

func TestAccRedTeamScanResource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: scan creation requires a validated target and takes a long time to complete.
	t.Skip("Red team scan creation requires a validated target and is long-running")
}
