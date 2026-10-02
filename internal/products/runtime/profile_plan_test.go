package runtime

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDLPReferenceMetadataUsesIdentity(t *testing.T) {
	ctx := context.Background()
	attributes := map[string]schema.Attribute{}
	for _, name := range []string{"name", "uuid", "profile_id", "version", "file_based"} {
		attributes[name] = schema.StringAttribute{Optional: true, Computed: true}
	}
	s := schema.Schema{Attributes: map[string]schema.Attribute{"entry": schema.SingleNestedAttribute{Optional: true, Attributes: attributes}}}
	raw := func(values map[string]any) tftypes.Value {
		fields := map[string]tftypes.Value{}
		for name := range attributes {
			fields[name] = tftypes.NewValue(tftypes.String, values[name])
		}
		kind := s.Type().TerraformType(ctx)
		entry := kind.(tftypes.Object).AttributeTypes["entry"]
		return tftypes.NewValue(kind, map[string]tftypes.Value{"entry": tftypes.NewValue(entry, fields)})
	}
	prior := map[string]any{"name": "first", "uuid": "uuid-first", "profile_id": "id-first", "version": "1", "file_based": "block"}
	for _, tc := range []struct {
		name     string
		config   map[string]any
		preserve bool
	}{
		{"same name", map[string]any{"name": "first"}, true},
		{"same UUID", map[string]any{"uuid": "uuid-first"}, true},
		{"same profile ID", map[string]any{"profile_id": "id-first"}, true},
		{"changed name", map[string]any{"name": "second"}, false},
		{"changed UUID or reordered entry", map[string]any{"uuid": "uuid-second"}, false},
		{"changed profile ID", map[string]any{"profile_id": "id-second"}, false},
		{"one identity changed", map[string]any{"uuid": "uuid-first", "profile_id": "id-second"}, false},
		{"changed version", map[string]any{"name": "first", "version": "2"}, false},
		{"unknown reference", map[string]any{"uuid": tftypes.UnknownValue}, false},
		{"no reference", map[string]any{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, leaf := range []string{"name", "uuid", "profile_id", "version", "file_based"} {
				request := planmodifier.StringRequest{Path: path.Root("entry").AtName(leaf), Config: tfsdk.Config{Raw: raw(tc.config), Schema: s}, State: tfsdk.State{Raw: raw(prior), Schema: s}, PlanValue: types.StringUnknown(), StateValue: types.StringValue(prior[leaf].(string))}
				response := planmodifier.StringResponse{PlanValue: request.PlanValue}
				sameDLPReferenceState().PlanModifyString(ctx, request, &response)
				if response.Diagnostics.HasError() {
					t.Fatal(response.Diagnostics)
				}
				if tc.preserve && !response.PlanValue.Equal(request.StateValue) {
					t.Errorf("%s lost stable metadata", leaf)
				}
				if !tc.preserve && !response.PlanValue.IsUnknown() {
					t.Errorf("%s inherited metadata from a different reference", leaf)
				}
			}
		})
	}
}
