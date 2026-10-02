package runtime

import (
	"context"
	"fmt"

	"github.com/cdot65/prisma-airs-provider/internal/tfutil"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Always request history explicitly. The service's default latest filter does
// not establish ownership or prove that all revisions have been deleted.
func namedProfileRevisions(ctx context.Context, client *airsruntime.Client, name string) ([]airsruntime.SecurityProfile, error) {
	latest := false
	var profiles []airsruntime.SecurityProfile
	seen := map[string]bool{}
	for offset := 0; ; {
		page, err := client.Profiles.ListWithOptions(ctx, airsruntime.ProfileListOpts{
			ListOpts: airsruntime.ListOpts{Limit: 100, Offset: offset}, Latest: &latest,
		})
		if err != nil {
			return nil, err
		}
		if len(page.Items) == 0 && page.NextOffset > offset {
			return nil, fmt.Errorf("profile pagination advanced without returning items")
		}
		newIDs := 0
		for _, profile := range page.Items {
			if profile.ProfileID == "" {
				return nil, fmt.Errorf("profile history omitted a revision UUID")
			}
			if !seen[profile.ProfileID] {
				newIDs++
				seen[profile.ProfileID] = true
			}
			if profile.ProfileName == name {
				profiles = append(profiles, profile)
			}
		}
		if len(page.Items) > 0 && newIDs == 0 {
			return nil, fmt.Errorf("profile pagination returned a repeated page")
		}
		if page.NextOffset > offset {
			offset = page.NextOffset
		} else if len(page.Items) >= 100 {
			offset += len(page.Items)
		} else {
			return profiles, nil
		}
	}
}

func latestNamedProfile(ctx context.Context, client *airsruntime.Client, name string) (*airsruntime.SecurityProfile, error) {
	profiles, err := namedProfileRevisions(ctx, client, name)
	if err != nil {
		return nil, err
	}
	var latest *airsruntime.SecurityProfile
	for i := range profiles {
		if latest == nil || profiles[i].Revision > latest.Revision {
			latest = &profiles[i]
		}
	}
	return latest, nil
}

func profileNameAvailable(ctx context.Context, client *airsruntime.Client, name string, diags *diag.Diagnostics) bool {
	existing, err := latestNamedProfile(ctx, client, name)
	if err != nil {
		diags.AddError("Failed to check security profile ownership", err.Error())
		return false
	}
	if existing != nil {
		profileOwnershipError(name, diags)
		return false
	}
	return true
}

func profileOwnershipError(name string, diags *diag.Diagnostics) {
	diags.AddError("Security profile name already exists", fmt.Sprintf(
		"Profile %q already exists. To manage it, explicitly import the profile name with terraform import <resource_address> %q. Creation and rename never adopt existing profiles.", name, name))
}

func profileOperationError(operation, name string, err error, diags *diag.Diagnostics) {
	if tfutil.IsConflict(err) {
		profileOwnershipError(name, diags)
		return
	}
	diags.AddError("Failed to "+operation+" security profile", err.Error())
}

func readCreatedProfile(ctx context.Context, client *airsruntime.Client, receipt *airsruntime.SecurityProfile, plan *SecurityProfileResourceModel, diags *diag.Diagnostics) *airsruntime.SecurityProfile {
	if receipt == nil || receipt.ProfileID == "" {
		diags.AddError("Missing security profile receipt", "The service did not return the new revision UUID; inspect the named profile before retrying.")
		return nil
	}
	// Record successful ownership before read-back, so a read failure does not
	// strand a remotely created revision outside Terraform state.
	plan.ID = types.StringValue(receipt.ProfileID)
	plan.ProfileID = plan.ID
	plan.Revision = types.Int64Value(int64(receipt.Revision))
	plan.Active = types.BoolValue(receipt.Active)
	plan.UpdatedAt = types.StringValue(receipt.LastModifiedTs)
	plan.CreatedAt = types.StringNull()
	profiles, err := namedProfileRevisions(ctx, client, plan.ProfileName.ValueString())
	if err != nil {
		diags.AddError("Failed to read new security profile revision", err.Error())
		return nil
	}
	for _, profile := range profiles {
		if profile.Revision == 1 {
			plan.CreatedAt = types.StringValue(profile.LastModifiedTs)
		}
	}

	for i := range profiles {
		if profiles[i].ProfileID == receipt.ProfileID {
			return &profiles[i]
		}
	}
	diags.AddError("New security profile revision not visible", "The write returned a UUID, but read-back did not find it. Refresh before retrying.")
	return nil
}
