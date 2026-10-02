package runtime

import (
	"context"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Computed reference identity stays known only while its configured name stays
// unchanged. A changed name must resolve a new identity, never inherit the old
// object's UUID merely because it occupies the same list position.
func sameNamedStringState(name string) planmodifier.String { return namedStringState{name} }

type namedStringState struct{ name string }

func (m namedStringState) Description(context.Context) string {
	return "Preserve known identity only while its configured name is unchanged."
}
func (m namedStringState) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }
func (m namedStringState) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	var planned, prior types.String
	path := req.Path.ParentPath().AtName(m.name)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path, &planned)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path, &prior)...)
	if !planned.IsUnknown() && planned.Equal(prior) {
		resp.PlanValue = req.StateValue
	}
}

func sameNamedInt64State(name string) planmodifier.Int64 { return namedInt64State{name} }

type namedInt64State struct{ name string }

func (m namedInt64State) Description(context.Context) string {
	return "Preserve known revision only while its configured name is unchanged."
}
func (m namedInt64State) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }
func (m namedInt64State) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	var planned, prior types.String
	path := req.Path.ParentPath().AtName(m.name)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path, &planned)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path, &prior)...)
	if !planned.IsUnknown() && planned.Equal(prior) {
		resp.PlanValue = req.StateValue
	}
}

func profileCreationTime(ctx context.Context, client *airsruntime.Client, name string, diags *diag.Diagnostics) types.String {
	profiles, err := namedProfileRevisions(ctx, client, name)
	if err != nil {
		diags.AddError("Failed to read profile creation metadata", err.Error())
		return types.StringNull()
	}
	for _, profile := range profiles {
		if profile.Revision == 1 {
			return types.StringValue(profile.LastModifiedTs)
		}
	}
	return types.StringNull()
}

func preservePartialProfileState(ctx context.Context, model *SecurityProfileResourceModel, state *tfsdk.State, diags *diag.Diagnostics) {
	diags.Append(state.Set(ctx, model)...)
	raw, err := tftypes.Transform(state.Raw, func(_ *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if !value.IsKnown() {
			return tftypes.NewValue(value.Type(), nil), nil
		}
		return value, nil
	})
	if err != nil {
		diags.AddError("Failed to preserve partial profile state", err.Error())
		return
	}
	state.Raw = raw
}

// DLP references may be configured by name or by service identity. Preserve
// computed reference metadata only when every configured identity/version still
// matches the prior entry; list position alone must never transfer metadata.
func sameDLPReferenceState() planmodifier.String { return dlpReferenceState{} }

type dlpReferenceState struct{}

func (dlpReferenceState) Description(context.Context) string {
	return "Preserve computed metadata only for an unchanged DLP reference."
}
func (m dlpReferenceState) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }
func (dlpReferenceState) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	identified := false
	for _, name := range []string{"name", "uuid", "profile_id", "version"} {
		var configured, prior types.String
		field := req.Path.ParentPath().AtName(name)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, field, &configured)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, field, &prior)...)
		if resp.Diagnostics.HasError() || configured.IsUnknown() {
			return
		}
		if configured.IsNull() {
			continue
		}
		if !configured.Equal(prior) {
			return
		}
		if name != "version" {
			identified = true
		}
	}
	if identified {
		resp.PlanValue = req.StateValue
	}
}
