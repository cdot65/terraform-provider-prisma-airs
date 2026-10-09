package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const profilePolicyPrivateKey = "security_profile_policy_v1"
const profileExtensionsPrivateKey = "security_profile_extensions_v1"

func saveProfileSnapshot(ctx context.Context, private profilePrivateState, profile *airsruntime.SecurityProfile, diags *diag.Diagnostics) {
	saveProfilePolicy(ctx, private, profile.Policy, diags)
	raw, err := json.Marshal(profile.Extensions)
	if err != nil {
		diags.AddError("Failed to preserve profile extensions", err.Error())
		return
	}
	diags.Append(private.SetKey(ctx, profileExtensionsPrivateKey, raw)...)
}
func preservedProfileExtensions(ctx context.Context, private profilePrivateState, diags *diag.Diagnostics) map[string]json.RawMessage {
	raw, d := private.GetKey(ctx, profileExtensionsPrivateKey)
	diags.Append(d...)
	if len(raw) == 0 {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		diags.AddError("Invalid preserved profile extensions", err.Error())
		return nil
	}
	return fields
}

type profilePrivateState interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

func saveProfilePolicy(ctx context.Context, private profilePrivateState, policy *airsruntime.ProfilePolicy, diags *diag.Diagnostics) {
	raw, err := json.Marshal(policy)
	if err != nil {
		diags.AddError("Failed to preserve security profile policy", err.Error())
		return
	}
	diags.Append(private.SetKey(ctx, profilePolicyPrivateKey, raw)...)
}

func preservedProfilePolicy(ctx context.Context, private profilePrivateState, state, plan *SecurityProfileResourceModel, diags *diag.Diagnostics) *airsruntime.ProfilePolicy {
	next := planToSDKPolicy(ctx, plan, diags)
	if next == nil {
		next = &airsruntime.ProfilePolicy{}
	}
	raw, d := private.GetKey(ctx, profilePolicyPrivateKey)
	diags.Append(d...)
	if len(raw) == 0 {
		diags.AddError("Security profile preservation unavailable", "Refresh this resource before updating it so the service policy and its extension fields can be preserved.")
		return nil
	}
	prior := planToSDKPolicy(ctx, state, diags)
	return mergeProfilePolicy(raw, prior, next, diags)
}

func policyJSON(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return decodePolicyJSON(raw)
}
func decodePolicyJSON(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	err := decoder.Decode(&value)
	return value, err
}
func mergeProfilePolicy(raw []byte, prior, next *airsruntime.ProfilePolicy, diags *diag.Diagnostics) *airsruntime.ProfilePolicy {
	original, err := decodePolicyJSON(raw)
	if err != nil {
		diags.AddError("Invalid preserved profile policy", err.Error())
		return nil
	}
	before, err := policyJSON(prior)
	if err != nil {
		diags.AddError("Invalid prior profile policy", err.Error())
		return nil
	}
	after, err := policyJSON(next)
	if err != nil {
		diags.AddError("Invalid planned profile policy", err.Error())
		return nil
	}
	merged, err := mergePolicyValue(original, before, after, "policy")
	if err != nil {
		diags.AddError("Cannot safely update security profile", err.Error())
		return nil
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		diags.AddError("Invalid merged profile policy", err.Error())
		return nil
	}
	var policy *airsruntime.ProfilePolicy
	if err = json.Unmarshal(encoded, &policy); err != nil {
		diags.AddError("Invalid merged profile policy", err.Error())
		return nil
	}
	return policy
}

// Apply changes to the Terraform projection, retaining raw service values where
// the projection did not change. Missing, null, empty and opaque fields therefore
// survive, while removal of a managed parent removes its complete subtree.
func mergePolicyValue(raw, before, after any, path string) (any, error) {
	if reflect.DeepEqual(before, after) {
		return raw, nil
	}
	if after == nil {
		return nil, nil
	}
	if next, ok := after.(map[string]any); ok {
		previous, _ := before.(map[string]any)
		original, _ := raw.(map[string]any)
		result := make(map[string]any, len(original)+len(next))
		for k, v := range original {
			result[k] = v
		}
		for k := range previous {
			if _, exists := next[k]; !exists {
				if err := validatePolicyRemoval(k, original[k], path); err != nil {
					return nil, err
				}
				delete(result, k)
			}
		}
		for k, v := range next {
			if prior, exists := previous[k]; exists {
				if _, present := original[k]; !present && reflect.DeepEqual(prior, v) {
					delete(result, k)
					continue
				}
				merged, err := mergePolicyValue(original[k], prior, v, path+"."+k)
				if err != nil {
					return nil, err
				}
				result[k] = merged
			} else {
				result[k] = v
			}
		}
		return result, nil
	}
	if next, ok := after.([]any); ok {
		if raw == nil {
			return next, nil
		}
		previous, _ := before.([]any)
		original, _ := raw.([]any)
		// Primitive arrays are wholly managed. Object lists need stable identities.
		if len(next) == 0 {
			return next, nil
		}
		if _, objects := next[0].(map[string]any); !objects {
			return next, nil
		}
		priorByID, err := policyEntries(previous, path)
		if err != nil {
			return nil, err
		}
		rawByID, err := policyEntries(original, path)
		if err != nil {
			return nil, err
		}
		nextByID, err := policyEntries(next, path)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(path, "model-protection") {
			for id, entry := range rawByID {
				if _, existed := priorByID[id]; !existed {
					continue
				}
				if _, remains := nextByID[id]; !remains {
					if err := validatePolicyRemoval("model-protection", []any{entry}, path); err != nil {
						return nil, err
					}
				}
			}
		}
		result := make([]any, 0, len(next))
		for _, entry := range next {
			id := policyEntryIdentity(entry.(map[string]any), path)
			if prior, exists := priorByID[id]; exists {
				saved, exists := rawByID[id]
				if !exists {
					return nil, fmt.Errorf("%s: no unambiguous service entry for %s; refresh before mutation", path, id)
				}
				merged, err := mergePolicyValue(saved, prior, entry, path+"["+id+"]")
				if err != nil {
					return nil, err
				}
				result = append(result, merged)
			} else {
				result = append(result, entry)
			}
		}
		return result, nil
	}
	return after, nil
}
func policyEntries(entries []any, path string) (map[string]any, error) {
	result := map[string]any{}
	for _, entry := range entries {
		object, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s contains an unsupported entry", path)
		}
		id := policyEntryIdentity(object, path)
		if _, exists := result[id]; exists {
			return nil, fmt.Errorf("%s has ambiguous repeated entry identity %s; mutation would risk transferring protection data", path, id)
		}
		result[id] = entry
	}
	return result, nil
}
func policyEntryIdentity(object map[string]any, path string) string {
	// Each supported list identifies entries within its owning profile/direction.
	fields := []string{"name"}
	switch {
	case strings.HasSuffix(path, "ai-security-profiles"):
		fields = []string{"model-type", "content-type"}
	case strings.HasSuffix(path, "dlp-data-profiles"):
		fields = []string{"name", "uuid", "id"}
	case strings.HasSuffix(path, "topic-list"):
		fields = []string{"action"}
	case strings.HasSuffix(path, "topic"):
		fields = []string{"topic_name", "topic_id"}
	case strings.HasSuffix(path, "toxic-category-list"):
		fields = []string{"category"}
	case strings.HasSuffix(path, "member"):
		fields = []string{"text"}
	}
	values := make([]string, len(fields))
	for i, field := range fields {
		if value, ok := object[field]; ok {
			values[i] = fmt.Sprint(value)
		}
	}
	raw, _ := json.Marshal(values)
	return string(raw)
}

// Fields which cannot be expressed by the typed schema must not disappear merely
// because a managed container or list was removed from configuration.
func validatePolicyRemoval(name string, raw any, path string) error {
	if name == "content-type-configurations" {
		if directions, ok := raw.(map[string]any); ok {
			for key := range directions {
				switch key {
				case "prompt", "response", "tool-call", "tool-response":
				default:
					return fmt.Errorf("%s contains unmodeled direction %q; keep the content_type_configurations block to preserve it before removing managed directions", path, key)
				}
			}
		}
	}
	if name == "model-protection" {
		if entries, ok := raw.([]any); ok {
			for _, entry := range entries {
				object, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				detector, _ := object["name"].(string)
				action, hasAction := object["action"].(string)
				categories, _ := object["toxic-category-list"].([]any)
				observedEmptyToxicAction := hasAction && detector == "toxic-content" && len(categories) > 0
				if detector == "" || !hasAction || (action == "" && !observedEmptyToxicAction) {
					return fmt.Errorf("%s contains a model detector without a configurable name/action; its removal cannot be represented safely", path)
				}
			}
		}
	}
	return nil
}
