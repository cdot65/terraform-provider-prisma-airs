package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func httpErr(status int) error {
	return aisec.NewHTTPError("boom", aisec.ClientSideError, status)
}

func undecodableErr() error {
	err := aisec.WrapError("failed to parse response JSON", aisec.AISecSDKInternalError, errors.New("bad json"))
	err.StatusCode = http.StatusOK
	return err
}

func TestErrorClassification(t *testing.T) {
	wrapped404 := fmt.Errorf("outer: %w", httpErr(http.StatusNotFound))

	if !isNotFound(httpErr(http.StatusNotFound)) || !isNotFound(wrapped404) {
		t.Error("404 (including wrapped) must be not-found")
	}
	if isNotFound(httpErr(http.StatusForbidden)) || isNotFound(errors.New("resource not found")) {
		t.Error("only typed 404 is not-found; text matching must not count")
	}
	if !isConflict(httpErr(http.StatusConflict)) || isConflict(httpErr(http.StatusNotFound)) {
		t.Error("isConflict must match only 409")
	}
	if !isUndecodableSuccess(undecodableErr()) {
		t.Error("undecodable success must be detected")
	}
	if isUndecodableSuccess(httpErr(http.StatusInternalServerError)) || isUndecodableSuccess(errors.New("failed to parse response JSON")) {
		t.Error("HTTP errors and untyped text must not be undecodable success")
	}
}

func TestFinishDelete(t *testing.T) {
	ctx := context.Background()
	absent := func(context.Context) (bool, error) { return true, nil }
	present := func(context.Context) (bool, error) { return false, nil }
	failing := func(context.Context) (bool, error) { return false, errors.New("list down") }

	cases := []struct {
		name    string
		err     error
		gone    func(context.Context) (bool, error)
		wantErr bool
	}{
		{"success", nil, nil, false},
		{"success verified", nil, absent, false},
		{"success still present", nil, present, true},
		{"success refresh fails", nil, failing, true},
		{"already gone 404", httpErr(404), nil, false},
		{"undecodable but gone", undecodableErr(), absent, false},
		{"undecodable still present", undecodableErr(), present, true},
		{"undecodable refresh fails", undecodableErr(), failing, true},
		{"undecodable without refresh", undecodableErr(), nil, true},
		{"server error", httpErr(500), absent, true},
		{"forbidden", httpErr(403), absent, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var diags diag.Diagnostics
			finishDelete(ctx, tc.err, "thing", tc.gone, &diags)
			if diags.HasError() != tc.wantErr {
				t.Fatalf("HasError=%v want %v: %v", diags.HasError(), tc.wantErr, diags)
			}
		})
	}
}
