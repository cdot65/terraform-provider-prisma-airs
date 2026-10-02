package supplychain

import (
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMapSecurityGroupToState_basic(t *testing.T) {
	group := &modelsecurity.ModelSecurityGroupResponse{
		UUID:        "sg-123",
		Name:        "test-group",
		Description: "Test security group",
		SourceType:  modelsecurity.SourceType("HUGGING_FACE"),
		State:       modelsecurity.ModelSecurityGroupState("ACTIVE"),
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-02T00:00:00Z",
	}

	var state ModelSecurityGroupResourceModel
	mapSecurityGroupToState(group, &state)

	assertStringValue(t, "ID", state.ID, "sg-123")
	assertStringValue(t, "UUID", state.UUID, "sg-123")
	assertStringValue(t, "Name", state.Name, "test-group")
	assertStringValue(t, "Description", state.Description, "Test security group")
	assertStringValue(t, "SourceType", state.SourceType, "HUGGING_FACE")
	assertStringValue(t, "State", state.State, "ACTIVE")
}

func assertStringValue(t *testing.T, field string, got types.String, want string) {
	t.Helper()
	if got.ValueString() != want {
		t.Errorf("%s: expected %q, got %q", field, want, got.ValueString())
	}
}
