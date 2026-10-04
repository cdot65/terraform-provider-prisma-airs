package runtime

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEmptyModelActionOnlyPreservesObservedToxicCategories(t *testing.T) {
	for _, tc := range []struct {
		name       string
		categories []ToxicCategoryModel
		valid      bool
	}{
		{"toxic-content", []ToxicCategoryModel{{Category: types.StringValue("harassment"), Action: types.StringValue("block")}}, true},
		{"toxic-content", nil, false}, {"prompt-injection", nil, false},
	} {
		model := SecurityProfileResourceModel{AiSecurityProfiles: []AiSecurityProfileModel{{ModelProtection: []ModelProtectionModel{{Name: types.StringValue(tc.name), Action: types.StringValue(""), ToxicCategories: tc.categories}}}}}
		var d diag.Diagnostics
		validateProfileActions(&model, &d)
		if d.HasError() == tc.valid {
			t.Fatalf("%s validity mismatch: %v", tc.name, d)
		}
	}
}
