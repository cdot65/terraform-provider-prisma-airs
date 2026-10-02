package redteam

import (
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMapPromptSetToState_basic(t *testing.T) {
	ps := &redteam.CustomPromptSetResponse{
		UUID:        "ps-123",
		Name:        "test-promptset",
		Description: "Test prompt set",
		Status:      "READY",
		Active:      true,
		Archive:     false,
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-02T00:00:00Z",
	}

	var state RedTeamCustomPromptSetResourceModel
	mapPromptSetToState(ps, &state)

	assertStringValue(t, "ID", state.ID, "ps-123")
	assertStringValue(t, "UUID", state.UUID, "ps-123")
	assertStringValue(t, "Name", state.Name, "test-promptset")
	assertStringValue(t, "Description", state.Description, "Test prompt set")
	assertStringValue(t, "Status", state.Status, "READY")
	assertBoolValue(t, "Active", state.Active, true)
	assertBoolValue(t, "Archive", state.Archive, false)
}

func assertStringValue(t *testing.T, field string, got types.String, want string) {
	t.Helper()
	if got.ValueString() != want {
		t.Errorf("%s: expected %q, got %q", field, want, got.ValueString())
	}
}

func assertBoolValue(t *testing.T, field string, got types.Bool, want bool) {
	t.Helper()
	if got.ValueBool() != want {
		t.Errorf("%s: expected %v, got %v", field, want, got.ValueBool())
	}
}
