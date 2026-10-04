package gateway

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestImportedNullMatchesNativeHCLAndRetainsDrift(t *testing.T) {
	ctx := context.Background()
	imported, err := readNative(ctx, map[string]any{"$eq": nil}, nil)
	if err != nil {
		t.Fatal(err)
	}
	configured := types.ObjectValueMust(map[string]attr.Type{"$eq": types.DynamicType}, map[string]attr.Value{"$eq": types.DynamicNull()})
	if !imported.Equal(configured) {
		t.Fatalf("imported null has different HCL type: %s", imported)
	}
	same, err := readNative(ctx, map[string]any{"$eq": nil}, imported)
	if err != nil || !same.Equal(configured) {
		t.Fatal("repeated refresh changed untyped null", err)
	}
	refreshed, err := readNative(ctx, map[string]any{"$eq": "changed"}, imported)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.Equal(configured) {
		t.Fatal("real null-to-string drift was suppressed")
	}
	typed, err := readNative(ctx, nil, types.StringNull())
	if err != nil || !typed.Equal(types.StringNull()) {
		t.Fatal("configured typed null was lost")
	}
}
