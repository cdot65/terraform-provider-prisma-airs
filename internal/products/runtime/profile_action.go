package runtime

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithValidateConfig = &securityProfileResource{}

func (r *securityProfileResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if !req.Config.Raw.IsFullyKnown() {
		return
	}
	var config SecurityProfileResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !resp.Diagnostics.HasError() {
		validateProfileActions(&config, &resp.Diagnostics)
	}
}

// Empty action is an observed API representation for toxic-content category
// settings. Preserve it exactly; it does not imply allow or block.
func validateProfileActions(model *SecurityProfileResourceModel, diags *diag.Diagnostics) {
	for _, profile := range model.AiSecurityProfiles {
		for _, protection := range profile.ModelProtection {
			if protection.Action.IsNull() || protection.Action.IsUnknown() || protection.Name.IsUnknown() {
				continue
			}
			if protection.Action.ValueString() == "" && (protection.Name.ValueString() != "toxic-content" || len(protection.ToxicCategories) == 0) {
				diags.AddError("Unsupported empty protection action", "An empty action is supported only for the observed toxic-content representation with explicit toxic_category settings. Preserve imported values; use a documented action for new rules.")
			}
		}
	}
}
