package supplychain

import (
	"errors"
	"fmt"
	"github.com/cdot65/prisma-airs-go/aisec"
)

// Upstream diagnostics can include finding content, authorization codes, and
// identities. Publish status/error class, never the raw API response message.
func skillError(err error) string {
	var sdk *aisec.AISecSDKError
	if errors.As(err, &sdk) {
		if sdk.StatusCode != 0 {
			return fmt.Sprintf("Skill Scanning API returned HTTP %d. Check authorization, endpoint settings, and service availability.", sdk.StatusCode)
		}
		return fmt.Sprintf("Skill Scanning SDK request failed (%s).", sdk.ErrorType)
	}
	return err.Error()
}

// Preserve typed error matching while preventing deletion helpers from echoing
// upstream bodies through their own diagnostics.
type skillRedactedError struct{ cause error }

func (e skillRedactedError) Error() string { return skillError(e.cause) }
func (e skillRedactedError) Unwrap() error { return e.cause }
func skillSafeError(err error) error {
	if err == nil {
		return nil
	}
	return skillRedactedError{cause: err}
}
func mutationMayHaveApplied(err error) bool {
	var sdk *aisec.AISecSDKError
	if errors.As(err, &sdk) {
		return sdk.StatusCode == 0 || sdk.StatusCode >= 500 || (sdk.StatusCode >= 200 && sdk.StatusCode < 300)
	}
	return err != nil
}
