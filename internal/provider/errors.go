package provider

import (
	"context"
	"errors"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// isNotFound reports whether err is an SDK 404 (server-side or a client-side
// lookup miss). Never match on error text.
func isNotFound(err error) bool {
	return aisec.IsNotFound(err)
}

// isConflict reports whether err is an SDK 409.
func isConflict(err error) bool {
	return errors.Is(err, aisec.ErrConflict)
}

// isUndecodableSuccess reports whether err is the SDK's "HTTP success but the
// body could not be decoded" failure. The remote operation may have completed,
// so callers must refresh before deciding whether to repeat or fail.
func isUndecodableSuccess(err error) bool {
	var sdkErr *aisec.AISecSDKError
	if !errors.As(err, &sdkErr) {
		return false
	}
	return (sdkErr.StatusCode >= 200 && sdkErr.StatusCode < 300) && sdkErr.ErrorType == aisec.AISecSDKInternalError
}

// finishDelete interprets the result of a delete call.
//
//   - nil error: success.
//   - 404: the object is already gone, which is the desired end state.
//   - undecodable success body: the delete may have been applied; call gone to
//     refresh and treat a confirmed absence as success.
//
// Anything else is reported as an error. gone may be nil when no refresh is
// possible, in which case undecodable bodies are reported as errors.
func finishDelete(ctx context.Context, err error, what string, gone func(context.Context) (bool, error), diags *diag.Diagnostics) {
	if err == nil || isNotFound(err) || isUndecodableSuccess(err) {
		if gone == nil {
			if err == nil || isNotFound(err) {
				return
			}
		} else {
			absent, checkErr := gone(ctx)
			if checkErr != nil {
				diags.AddError("Failed to verify deletion of "+what, checkErr.Error())
				return
			}
			if absent {
				tflog.Debug(ctx, "delete or archive confirmed", map[string]any{"resource": what})
				return
			}
			diags.AddError("Deletion of "+what+" not confirmed", "The service still reports the object as present. Retry destroy after checking its lifecycle status.")
			return
		}
	}

	diags.AddError("Failed to delete "+what, err.Error())
}
