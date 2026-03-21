package provider_test

import (
	"testing"
)

func TestAccModelScanResource_basic(t *testing.T) {
	testAccPreCheck(t)
	// Skip: Model scan creation requires a valid model source (e.g. HuggingFace model name)
	// and an active security group. The scan process is long-running and may not complete
	// within test timeouts.
	t.Skip("Model scan creation requires active security group and valid model source")
}
