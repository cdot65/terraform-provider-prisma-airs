package runtime

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func directionalFixture(t *testing.T) airsruntime.SecurityProfile {
	t.Helper()
	raw, err := os.ReadFile("testdata/directional-security-profile.json")
	if err != nil {
		t.Fatal(err)
	}
	var profile airsruntime.SecurityProfile
	if err = json.Unmarshal(raw, &profile); err != nil {
		t.Fatal(err)
	}
	return profile
}
func assertPolicyJSON(t *testing.T, want, got any) {
	t.Helper()
	expected, err := policyJSON(want)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := policyJSON(got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, actual) {
		a, _ := json.Marshal(expected)
		b, _ := json.Marshal(actual)
		t.Fatalf("policy changed\nwant %s\ngot  %s", a, b)
	}
}
func TestDirectionalProfileReadEditPreservesPolicy(t *testing.T) {
	ctx := context.Background()
	profile := directionalFixture(t)
	if err := profile.Policy.SetExtension("future-policy", json.RawMessage(`{"large":9007199254740993}`)); err != nil {
		t.Fatal(err)
	}
	directions := profile.Policy.AiSecurityProfiles[0].ContentTypeConfigurations
	if err := directions.SetExtension("future-direction", json.RawMessage(`{"enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	if err := directions.Response.ModelProtection[0].SetExtension("future-detector", json.RawMessage(`{"enabled":false}`)); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(profile.Policy)
	if err != nil {
		t.Fatal(err)
	}
	var state SecurityProfileResourceModel
	var diags diag.Diagnostics
	mapProfileToState(ctx, &profile, &state, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	p := state.AiSecurityProfiles[0]
	if p.EnableFullConversationInspection.IsNull() || p.EnableFullConversationInspection.ValueBool() {
		t.Fatal("explicit conversation false lost")
	}
	if !p.ContentTypeConfigurations.ToolResponse.DataProtection.DataLeakDetection.MaskDataInline.IsNull() {
		t.Fatal("omitted inline masking became false")
	}
	if id := p.ContentTypeConfigurations.Prompt.DataProtection.DataLeakDetection.Members[0].ID; id.IsNull() || id.ValueString() != "" {
		t.Fatal("explicit empty ID lost")
	}
	prior := planToSDKPolicy(ctx, &state, &diags)
	// Rehydrate raw bytes as another Terraform process would.
	var reloaded airsruntime.ProfilePolicy
	if err = json.Unmarshal(raw, &reloaded); err != nil {
		t.Fatal(err)
	}
	persisted, err := json.Marshal(reloaded)
	if err != nil {
		t.Fatal(err)
	}
	unchanged := mergeProfilePolicy(persisted, prior, prior, &diags)
	assertPolicyJSON(t, profile.Policy, unchanged)
	state.AiSecurityProfiles[0].ContentTypeConfigurations.Response.ModelProtection[0].SeverityByConfidence.High = types.StringValue("critical")
	next := planToSDKPolicy(ctx, &state, &diags)
	merged := mergeProfilePolicy(persisted, prior, next, &diags)
	directions.Response.ModelProtection[0].SeverityByConfidence.High = "critical"
	if diags.HasError() {
		t.Fatal(diags)
	}
	assertPolicyJSON(t, profile.Policy, merged)
	// A global-only edit also retains every direction and all opaque values.
	raw, err = json.Marshal(merged)
	if err != nil {
		t.Fatal(err)
	}
	prior = next
	state.AiSecurityProfiles[0].Latency.MaxInlineLatency = types.Int64Value(9)
	next = planToSDKPolicy(ctx, &state, &diags)
	merged = mergeProfilePolicy(raw, prior, next, &diags)
	profile.Policy.AiSecurityProfiles[0].ModelConfiguration.Latency.MaxInlineLatency = 9
	assertPolicyJSON(t, profile.Policy, merged)
}
func TestDirectionalManagedDetectorRemovalAndReorder(t *testing.T) {
	ctx := context.Background()
	profile := directionalFixture(t)
	raw, err := json.Marshal(profile.Policy)
	if err != nil {
		t.Fatal(err)
	}
	var state SecurityProfileResourceModel
	var diags diag.Diagnostics
	mapProfileToState(ctx, &profile, &state, &diags)
	prior := planToSDKPolicy(ctx, &state, &diags)
	state.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection = state.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection[1:]
	next := planToSDKPolicy(ctx, &state, &diags)
	merged := mergeProfilePolicy(raw, prior, next, &diags)
	profile.Policy.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection = profile.Policy.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection[1:]
	if diags.HasError() {
		t.Fatal(diags)
	}
	assertPolicyJSON(t, profile.Policy, merged)
	var refreshed SecurityProfileResourceModel
	profile.Policy = merged
	mapProfileToState(ctx, &profile, &refreshed, &diags)
	if len(refreshed.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection) != 1 {
		t.Fatal("removed detector resurrected")
	}
	// Reordered detectors must carry their own extension metadata.
	profile = directionalFixture(t)
	prompt := profile.Policy.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt
	if err = prompt.ModelProtection[0].SetExtension("marker", json.RawMessage(`"injection"`)); err != nil {
		t.Fatal(err)
	}
	if err = prompt.ModelProtection[1].SetExtension("marker", json.RawMessage(`"toxicity"`)); err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(profile.Policy)
	if err != nil {
		t.Fatal(err)
	}
	mapProfileToState(ctx, &profile, &state, &diags)
	prior = planToSDKPolicy(ctx, &state, &diags)
	models := state.AiSecurityProfiles[0].ContentTypeConfigurations.Prompt.ModelProtection
	models[0], models[1] = models[1], models[0]
	merged = mergeProfilePolicy(raw, prior, planToSDKPolicy(ctx, &state, &diags), &diags)
	prompt.ModelProtection[0], prompt.ModelProtection[1] = prompt.ModelProtection[1], prompt.ModelProtection[0]
	assertPolicyJSON(t, profile.Policy, merged)
}
func TestProtectionBooleanPresence(t *testing.T) {
	for _, value := range []types.Bool{types.BoolNull(), types.BoolUnknown(), types.BoolValue(false), types.BoolValue(true)} {
		m := &ProtectionModel{DataProtection: &DataProtectionModel{DataLeakDetection: &DataLeakDetectionModel{Action: types.StringValue("block"), MaskDataInline: value}}}
		var diags diag.Diagnostics
		sdk := protectionToSDK(context.Background(), m, &diags).DataProtection.DataLeakDetection
		raw, err := json.Marshal(sdk)
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]json.RawMessage
		if err = json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		_, present := object["mask-data-inline"]
		if present == (value.IsNull() || value.IsUnknown()) {
			t.Fatalf("incorrect presence for %v: %s", value, raw)
		}
	}
}
func TestTopicsDoNotBorrowAcrossProfilesOrRestoreEmpty(t *testing.T) {
	prior := []AiSecurityProfileModel{{ModelType: types.StringValue("first"), ModelProtection: []ModelProtectionModel{{Name: types.StringValue("topic-guardrails"), TopicLists: []TopicListModel{{Action: types.StringValue("block"), Topics: []TopicRefModel{{TopicName: types.StringValue("secret"), TopicID: types.StringValue("id"), Revision: types.Int64Value(1)}}}}}}}}
	profile := &airsruntime.SecurityProfile{Policy: &airsruntime.ProfilePolicy{AiSecurityProfiles: []airsruntime.AiSecurityProfileConfig{{ModelType: "second", ModelConfiguration: &airsruntime.ModelConfiguration{ModelProtection: []airsruntime.ModelProtectionConfig{{Name: "topic-guardrails", TopicList: []airsruntime.TopicArrayConfig{{Action: "block"}}}}}}}}}
	state := SecurityProfileResourceModel{AiSecurityProfiles: prior}
	var diags diag.Diagnostics
	mapProfileToState(context.Background(), profile, &state, &diags)
	if len(state.AiSecurityProfiles[0].ModelProtection[0].TopicLists[0].Topics) != 0 {
		t.Fatal("topics borrowed from another profile")
	}
	profile.Policy.AiSecurityProfiles[0].ModelType = "first"
	profile.Policy.AiSecurityProfiles[0].ModelConfiguration.ModelProtection[0].TopicList[0].Topic = []airsruntime.TopicRef{}
	state.AiSecurityProfiles = prior
	mapProfileToState(context.Background(), profile, &state, &diags)
	if len(state.AiSecurityProfiles[0].ModelProtection[0].TopicLists[0].Topics) != 0 {
		t.Fatal("empty topics resurrected")
	}
}
func TestAmbiguousPolicyMergeFailsBeforeMutation(t *testing.T) {
	before := map[string]any{"model-protection": []any{map[string]any{"name": "duplicate", "action": "block"}, map[string]any{"name": "duplicate", "action": "allow"}}}
	after := map[string]any{"model-protection": []any{map[string]any{"name": "duplicate", "action": "block"}}}
	if _, err := mergePolicyValue(before, before, after, "policy"); err == nil {
		t.Fatal("ambiguous identity accepted")
	}
}

func TestLegacyAndDirectionalProtectionExtensions(t *testing.T) {
	for _, directional := range []bool{false, true} {
		protection := `{"data-protection":{"source-code-detection":{"action":"","severity":""}},"app-protection":{"url-detected-severity":"","allow-url-category":{"member":[]}},"model-protection":[{"name":"future-model","action":"observe","severity":"","severity-by-confidence":{"high":"custom","moderate":""},"toxic-category-list":[{"category":"violence","action":"high:block, moderate:allow","severity-by-confidence":{"high":"high"}}],"topic-list":[{"action":"allow","topic":[{"topic_name":"topic","topic_id":"topic-id","revision":2,"severity":"medium"}]}],"options":[{"precision":9007199254740993}]}],"agent-protection":[{"name":"future-agent","action":"observe","severity":"custom"}]}`
		raw := `{"model-type":"default","model-configuration":` + protection + `}`
		if directional {
			raw = `{"model-type":"default","content-type-mode":"future-mode","content-type-configurations":{"response":` + protection + `}}`
		}
		var asp airsruntime.AiSecurityProfileConfig
		if err := json.Unmarshal([]byte(raw), &asp); err != nil {
			t.Fatal(err)
		}
		profile := airsruntime.SecurityProfile{Policy: &airsruntime.ProfilePolicy{AiSecurityProfiles: []airsruntime.AiSecurityProfileConfig{asp}}}
		saved, err := json.Marshal(profile.Policy)
		if err != nil {
			t.Fatal(err)
		}
		var state SecurityProfileResourceModel
		var diags diag.Diagnostics
		mapProfileToState(context.Background(), &profile, &state, &diags)
		m := state.AiSecurityProfiles[0].protection()
		if directional {
			m = state.AiSecurityProfiles[0].ContentTypeConfigurations.Response
		}
		if m.ModelProtection[0].Severity.IsNull() || m.DataProtection.SourceCodeDetection.Severity.IsNull() || !m.AppProtection.AllowURLCategory.Equal(types.ListValueMust(types.StringType, []attr.Value{})) {
			t.Fatal("empty values collapsed")
		}
		prior := planToSDKPolicy(context.Background(), &state, &diags)
		m.ModelProtection[0].Action = types.StringValue("block")
		next := planToSDKPolicy(context.Background(), &state, &diags)
		merged := mergeProfilePolicy(saved, prior, next, &diags)
		if directional {
			asp.ContentTypeConfigurations.Response.ModelProtection[0].Action = "block"
		} else {
			asp.ModelConfiguration.ModelProtection[0].Action = "block"
		}
		if diags.HasError() {
			t.Fatal(diags)
		}
		assertPolicyJSON(t, profile.Policy, merged)
	}
}

func TestUnrepresentablePolicyRemovalIsRejected(t *testing.T) {
	for _, before := range []map[string]any{
		{"content-type-configurations": map[string]any{"future-direction": map[string]any{"enabled": true}}},
		{"model-protection": []any{map[string]any{"options": []any{"unrepresented"}}}},
	} {
		if _, err := mergePolicyValue(before, before, map[string]any{}, "policy"); err == nil {
			t.Fatal("unrepresented policy data removed")
		}
	}
}

func TestPartialRemovalOfUnrepresentableDetectorIsRejected(t *testing.T) {
	before := map[string]any{"model-protection": []any{map[string]any{"options": []any{"opaque"}}, map[string]any{"name": "prompt-injection", "action": "block"}}}
	after := map[string]any{"model-protection": []any{map[string]any{"name": "prompt-injection", "action": "block"}}}
	if _, err := mergePolicyValue(before, before, after, "policy"); err == nil {
		t.Fatal("unrepresentable detector silently removed from a partially retained list")
	}
}
