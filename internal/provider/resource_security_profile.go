package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &securityProfileResource{}
	_ resource.ResourceWithImportState = &securityProfileResource{}
)

func NewSecurityProfileResource() resource.Resource {
	return &securityProfileResource{}
}

type securityProfileResource struct {
	client *management.Client
}

type SecurityProfileResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProfileID   types.String `tfsdk:"profile_id"`
	ProfileName types.String `tfsdk:"profile_name"`
	Policy      types.String `tfsdk:"policy"`
	Active      types.Bool   `tfsdk:"active"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *securityProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_profile"
}

func (r *securityProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AI Runtime Security profile.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform resource ID (same as profile_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"profile_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the security profile.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"profile_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the security profile.",
			},
			"policy": schema.StringAttribute{
				Optional:    true,
				Description: "Policy configuration as a JSON string.",
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the profile is active.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp.",
			},
		},
	}
}

func (r *securityProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getMgmtClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = client
}

func (r *securityProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := management.CreateProfileRequest{
		ProfileName: plan.ProfileName.ValueString(),
	}

	if !plan.Policy.IsNull() && !plan.Policy.IsUnknown() && plan.Policy.ValueString() != "" {
		var policy management.ProfilePolicy
		if err := json.Unmarshal([]byte(plan.Policy.ValueString()), &policy); err != nil {
			resp.Diagnostics.AddError("Invalid policy JSON", err.Error())
			return
		}
		createReq.Policy = &policy
	}

	profile, err := r.client.Profiles.Create(ctx, createReq)
	if err != nil {
		// The API may return 409 even when the profile was actually created
		// (known API bug with certain policy configurations like toxic-category-list).
		// Attempt to read back the profile by name before failing.
		if strings.Contains(err.Error(), "409") {
			found := findProfileByName(ctx, r.client, plan.ProfileName.ValueString())
			if found != nil {
				tflog.Warn(ctx, "profile create returned 409 but profile exists; treating as success")
				mapProfileToState(found, &plan)
				resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
				return
			}
		}
		resp.Diagnostics.AddError("Failed to create security profile", err.Error())
		return
	}

	mapProfileToState(profile, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := findProfileByID(ctx, r.client, state.ProfileID.ValueString())
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapProfileToState(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := management.UpdateProfileRequest{
		ProfileName: plan.ProfileName.ValueString(),
	}

	if !plan.Policy.IsNull() && !plan.Policy.IsUnknown() && plan.Policy.ValueString() != "" {
		var policy management.ProfilePolicy
		if err := json.Unmarshal([]byte(plan.Policy.ValueString()), &policy); err != nil {
			resp.Diagnostics.AddError("Invalid policy JSON", err.Error())
			return
		}
		updateReq.Policy = &policy
	}

	profile, err := r.client.Profiles.Update(ctx, state.ProfileID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update security profile", err.Error())
		return
	}

	mapProfileToState(profile, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Profiles.Delete(ctx, state.ProfileID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "failed to parse response JSON") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete security profile", err.Error())
		return
	}
}

func (r *securityProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	profileName := req.ID

	found := findProfileByName(ctx, r.client, profileName)
	if found == nil {
		resp.Diagnostics.AddError("Profile not found", "No profile with name: "+profileName)
		return
	}

	var state SecurityProfileResourceModel
	mapProfileToState(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// findProfileByID searches for a profile by ID using paginated list calls.
func findProfileByID(ctx context.Context, client *management.Client, profileID string) *management.SecurityProfile {
	offset := 0
	limit := 100
	for {
		listResp, err := client.Profiles.List(ctx, management.ListOpts{Limit: limit, Offset: offset})
		if err != nil {
			return nil
		}
		for i := range listResp.Items {
			if listResp.Items[i].ProfileID == profileID {
				return &listResp.Items[i]
			}
		}
		if len(listResp.Items) < limit {
			return nil
		}
		offset += limit
	}
}

// findProfileByName searches for a profile by name using paginated list calls.
func findProfileByName(ctx context.Context, client *management.Client, profileName string) *management.SecurityProfile {
	offset := 0
	limit := 100
	for {
		listResp, err := client.Profiles.List(ctx, management.ListOpts{Limit: limit, Offset: offset})
		if err != nil {
			return nil
		}
		for i := range listResp.Items {
			if listResp.Items[i].ProfileName == profileName {
				return &listResp.Items[i]
			}
		}
		if len(listResp.Items) < limit {
			return nil
		}
		offset += limit
	}
}

func mapProfileToState(profile *management.SecurityProfile, state *SecurityProfileResourceModel) {
	state.ID = types.StringValue(profile.ProfileID)
	state.ProfileID = types.StringValue(profile.ProfileID)
	state.ProfileName = types.StringValue(profile.ProfileName)
	state.Active = types.BoolValue(profile.Active)
	state.CreatedAt = types.StringValue(profile.LastModifiedTs)
	state.UpdatedAt = types.StringValue(profile.LastModifiedTs)

	if profile.Policy != nil {
		policyJSON, err := json.Marshal(profile.Policy)
		if err == nil {
			newPolicyStr := string(policyJSON)
			// Preserve the original JSON string if semantically equivalent
			// to avoid Terraform detecting a change due to key ordering.
			if !state.Policy.IsNull() && !state.Policy.IsUnknown() {
				var existing, returned any
				if json.Unmarshal([]byte(state.Policy.ValueString()), &existing) == nil &&
					json.Unmarshal(policyJSON, &returned) == nil {
					existingNorm, _ := json.Marshal(existing)
					returnedNorm, _ := json.Marshal(returned)
					if string(existingNorm) == string(returnedNorm) {
						return // keep existing policy string
					}
				}
			}
			state.Policy = types.StringValue(newPolicyStr)
		}
	}
}

func getMgmtClient(data any) (*management.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	pd, ok := data.(*ProviderData)
	if !ok || pd == nil {
		diags.AddError("Missing provider config", "Provider not configured")
		return nil, diags
	}
	if pd.MgmtClient == nil {
		diags.AddError("Management client not configured", "OAuth2 credentials required")
		return nil, diags
	}
	return pd.MgmtClient, diags
}
