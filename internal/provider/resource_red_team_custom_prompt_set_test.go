package provider_test

import (
	"testing"
)

func TestAccRedTeamCustomPromptSetResource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: Custom prompt set API hangs with current service account.
	// The API endpoint does not respond properly, causing test timeout.
	t.Skip("Custom prompt set API hangs with current service account permissions")
}
