package runtime

import "github.com/hashicorp/terraform-plugin-framework/diag"

// Empty action is an observed API representation for toxic-content category
// settings. Preserve it exactly; it does not imply allow or block.
func validateProfileActions(model *SecurityProfileResourceModel, diags *diag.Diagnostics) {
	for _, profile := range model.AiSecurityProfiles {
		protections := []*ProtectionModel{profile.protection()}
		if directions := profile.ContentTypeConfigurations; directions != nil {
			protections = append(protections, directions.Prompt, directions.Response, directions.ToolCall, directions.ToolResponse)
		}
		for _, layout := range protections {
			if layout == nil {
				continue
			}
			for _, protection := range layout.ModelProtection {
				if protection.Action.IsNull() || protection.Action.IsUnknown() || protection.Name.IsUnknown() {
					continue
				}
				if protection.Action.ValueString() == "" && (protection.Name.ValueString() != "toxic-content" || len(protection.ToxicCategories) == 0) {
					diags.AddError("Unsupported empty protection action", "An empty action is supported only for the observed toxic-content representation with explicit toxic_category settings. Preserve imported values; use a documented action for new rules.")
				}
			}
		}
	}
}
