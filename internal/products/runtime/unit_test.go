package runtime

import (
	"context"
	"testing"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mapApiKeyToState

func TestMapApiKeyToState_basic(t *testing.T) {
	key := &airsruntime.ApiKey{
		ApiKeyID:   "key-123",
		ApiKeyName: "my-key",
		ApiKey:     "secret-value",
		Revoked:    false,
		Status:     "active",
		CreationTS: "2026-01-01T00:00:00Z",
		Expiration: "2027-01-01T00:00:00Z",
	}

	var state ApiKeyResourceModel
	mapApiKeyToState(key, &state)

	assertStringValue(t, "ID", state.ID, "key-123")
	assertStringValue(t, "ApiKeyID", state.ApiKeyID, "key-123")
	assertStringValue(t, "ApiKeyName", state.ApiKeyName, "my-key")
	assertStringValue(t, "ApiKey", state.ApiKey, "secret-value")
	assertBoolValue(t, "Revoked", state.Revoked, false)
	assertStringValue(t, "Status", state.Status, "active")
	assertStringValue(t, "CreatedAt", state.CreatedAt, "2026-01-01T00:00:00Z")
	assertStringValue(t, "ExpiresAt", state.ExpiresAt, "2027-01-01T00:00:00Z")
}

func TestMapApiKeyToState_emptyApiKey(t *testing.T) {
	key := &airsruntime.ApiKey{
		ApiKeyID:   "key-456",
		ApiKeyName: "another-key",
		ApiKey:     "", // empty — should not overwrite
		Revoked:    true,
		Status:     "revoked",
		CreationTS: "2026-02-01T00:00:00Z",
		Expiration: "2027-06-15T12:00:00Z",
	}

	state := ApiKeyResourceModel{
		ApiKey: types.StringValue("previous-secret"),
	}
	mapApiKeyToState(key, &state)

	// empty ApiKey should NOT overwrite existing state value
	assertStringValue(t, "ApiKey", state.ApiKey, "previous-secret")
	assertBoolValue(t, "Revoked", state.Revoked, true)
	assertStringValue(t, "Status", state.Status, "revoked")
}

// mapAppToState

func TestMapAppToState_basic(t *testing.T) {
	app := &airsruntime.CustomerApp{
		CustomerAppID:    "app-789",
		AppName:          "test-app",
		TsgID:            "tsg-001",
		ModelName:        "gpt-4",
		CloudProvider:    "aws",
		Environment:      "production",
		Status:           "active",
		CreatedBy:        "user@example.com",
		AgentApp:         true,
		AiSecProfileName: "my-profile",
	}

	var state CustomerAppResourceModel
	mapAppToState(app, &state)

	assertStringValue(t, "ID", state.ID, "app-789")
	assertStringValue(t, "CustomerAppID", state.CustomerAppID, "app-789")
	assertStringValue(t, "AppName", state.AppName, "test-app")
	assertStringValue(t, "TsgID", state.TsgID, "tsg-001")
	assertStringValue(t, "ModelName", state.ModelName, "gpt-4")
	assertStringValue(t, "CloudProvider", state.CloudProvider, "aws")
	assertStringValue(t, "Environment", state.Environment, "production")
	assertStringValue(t, "Status", state.Status, "active")
	assertStringValue(t, "CreatedBy", state.CreatedBy, "user@example.com")
	assertStringValue(t, "AiSecProfileName", state.AiSecProfileName, "my-profile")
	if !state.AgentApp.ValueBool() {
		t.Errorf("AgentApp = false, want true")
	}
}

func TestMapAppToState_emptyFields(t *testing.T) {
	app := &airsruntime.CustomerApp{
		CustomerAppID: "app-000",
	}

	var state CustomerAppResourceModel
	mapAppToState(app, &state)

	assertStringValue(t, "ID", state.ID, "app-000")
	assertStringValue(t, "AppName", state.AppName, "")
}

// Schema validator tests (TDD: should fail until validators added)

func TestSecurityProfileSchema_ModelProtectionValidators(t *testing.T) {
	r := &securityProfileResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	aspBlock := resp.Schema.Blocks["ai_security_profile"].(schema.ListNestedBlock)
	mpBlock := aspBlock.NestedObject.Blocks["model_protection"].(schema.ListNestedBlock)

	nameAttr := mpBlock.NestedObject.Attributes["name"].(schema.StringAttribute)
	if len(nameAttr.Validators) == 0 {
		t.Error("model_protection.name: expected validators, got none")
	}

	actionAttr := mpBlock.NestedObject.Attributes["action"].(schema.StringAttribute)
	// Empty toxicity actions require sibling category context, so resource-level
	// validation enforces them. Future nonempty action strings remain accepted.
	if !actionAttr.Required {
		t.Error("model_protection.action: expected a required attribute")
	}
}

func TestSecurityProfileSchema_AgentProtectionValidators(t *testing.T) {
	r := &securityProfileResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	aspBlock := resp.Schema.Blocks["ai_security_profile"].(schema.ListNestedBlock)
	apBlock := aspBlock.NestedObject.Blocks["agent_protection"].(schema.ListNestedBlock)

	nameAttr := apBlock.NestedObject.Attributes["name"].(schema.StringAttribute)
	if len(nameAttr.Validators) == 0 {
		t.Error("agent_protection.name: expected validators, got none")
	}

	actionAttr := apBlock.NestedObject.Attributes["action"].(schema.StringAttribute)
	if len(actionAttr.Validators) == 0 {
		t.Error("agent_protection.action: expected validators, got none")
	}
}

func TestSecurityProfileSchema_LatencyValidators(t *testing.T) {
	r := &securityProfileResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	aspBlock := resp.Schema.Blocks["ai_security_profile"].(schema.ListNestedBlock)
	latencyBlock := aspBlock.NestedObject.Blocks["latency"].(schema.SingleNestedBlock)

	actionAttr := latencyBlock.Attributes["inline_timeout_action"].(schema.StringAttribute)
	if len(actionAttr.Validators) == 0 {
		t.Error("latency.inline_timeout_action: expected validators, got none")
	}
}

func TestSecurityProfileSchema_NoAlertInDescriptions(t *testing.T) {
	r := &securityProfileResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	aspBlock := resp.Schema.Blocks["ai_security_profile"].(schema.ListNestedBlock)
	mpBlock := aspBlock.NestedObject.Blocks["model_protection"].(schema.ListNestedBlock)

	actionAttr := mpBlock.NestedObject.Attributes["action"].(schema.StringAttribute)
	if containsSubstring(actionAttr.Description, "alert") {
		t.Errorf("model_protection.action description should not mention 'alert', got: %s", actionAttr.Description)
	}
}

// ToxicContentAction compound value tests

func TestPlanToSDKPolicy_compoundToxicContentAction(t *testing.T) {
	ctx := context.Background()
	plan := &SecurityProfileResourceModel{
		AiSecurityProfiles: []AiSecurityProfileModel{
			{
				ModelType: types.StringValue("default"),
				ModelProtection: []ModelProtectionModel{
					{
						Name:   types.StringValue("toxic-content"),
						Action: types.StringValue("high:block, moderate:allow"),
					},
				},
			},
		},
	}

	var diags diag.Diagnostics
	policy := planToSDKPolicy(ctx, plan, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if policy == nil {
		t.Fatal("expected non-nil policy")
	}

	mp := policy.AiSecurityProfiles[0].ModelConfiguration.ModelProtection[0]
	if mp.Name != "toxic-content" {
		t.Errorf("expected 'toxic-content', got %q", mp.Name)
	}
	if string(mp.Action) != "high:block, moderate:allow" {
		t.Errorf("expected 'high:block, moderate:allow', got %q", string(mp.Action))
	}
}

func TestMapProfileToState_compoundToxicContentAction(t *testing.T) {
	ctx := context.Background()
	profile := &airsruntime.SecurityProfile{
		ProfileID:      "prof-tc",
		ProfileName:    "toxic-compound",
		Active:         true,
		LastModifiedTs: "2026-03-01T00:00:00Z",
		Policy: &airsruntime.ProfilePolicy{
			AiSecurityProfiles: []airsruntime.AiSecurityProfileConfig{
				{
					ModelType: "default",
					ModelConfiguration: &airsruntime.ModelConfiguration{
						ModelProtection: []airsruntime.ModelProtectionConfig{
							{
								Name:   "toxic-content",
								Action: "high:block, moderate:block",
							},
						},
					},
				},
			},
		},
	}

	var state SecurityProfileResourceModel
	var diags diag.Diagnostics
	mapProfileToState(ctx, profile, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	mp := state.AiSecurityProfiles[0].ModelProtection[0]
	assertStringValue(t, "Name", mp.Name, "toxic-content")
	assertStringValue(t, "Action", mp.Action, "high:block, moderate:block")
}

// mapProfileToState

func TestMapProfileToState_basic(t *testing.T) {
	ctx := context.Background()
	profile := &airsruntime.SecurityProfile{
		ProfileID:      "prof-123",
		ProfileName:    "default",
		Active:         true,
		Policy:         &airsruntime.ProfilePolicy{},
		LastModifiedTs: "2026-01-02T00:00:00Z",
	}

	var state SecurityProfileResourceModel
	var diags diag.Diagnostics
	mapProfileToState(ctx, profile, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	assertStringValue(t, "ID", state.ID, "prof-123")
	assertStringValue(t, "ProfileID", state.ProfileID, "prof-123")
	assertStringValue(t, "ProfileName", state.ProfileName, "default")
	assertBoolValue(t, "Active", state.Active, true)
	if !state.CreatedAt.IsNull() {
		t.Error("creation metadata must come from revision 1 history, not latest modification time")
	}
	assertStringValue(t, "UpdatedAt", state.UpdatedAt, "2026-01-02T00:00:00Z")
}

func TestMapProfileToState_nilPolicy(t *testing.T) {
	ctx := context.Background()
	profile := &airsruntime.SecurityProfile{
		ProfileID:   "prof-456",
		ProfileName: "minimal",
		Policy:      nil,
	}

	var state SecurityProfileResourceModel
	var diags diag.Diagnostics
	mapProfileToState(ctx, profile, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	assertStringValue(t, "ID", state.ID, "prof-456")
	if state.AiSecurityProfiles != nil {
		t.Error("AiSecurityProfiles: expected nil for nil policy")
	}
	if state.DlpDataProfiles != nil {
		t.Error("DlpDataProfiles: expected nil for nil policy")
	}
}

func TestMapProfileToState_fullPolicy(t *testing.T) {
	ctx := context.Background()
	profile := &airsruntime.SecurityProfile{
		ProfileID:      "prof-789",
		ProfileName:    "full-test",
		Active:         true,
		LastModifiedTs: "2026-03-01T00:00:00Z",
		Policy: &airsruntime.ProfilePolicy{
			AiSecurityProfiles: []airsruntime.AiSecurityProfileConfig{
				{
					ModelType: "default",
					ModelConfiguration: &airsruntime.ModelConfiguration{
						MaskDataInStorage: true,
						Latency: &airsruntime.LatencyConfig{
							InlineTimeoutAction: "allow",
							MaxInlineLatency:    30,
						},
						ModelProtection: []airsruntime.ModelProtectionConfig{
							{
								Name:   "prompt-injection",
								Action: "block",
							},
							{
								Name:   "toxic-content",
								Action: "high:block, moderate:allow",
								ToxicCategoryList: []airsruntime.ToxicCategoryConfig{
									{Category: "harassment", Action: "block"},
									{Category: "violence", Action: "block"},
								},
							},
						},
						AgentProtection: []airsruntime.AgentProtectionConfig{
							{Name: "agent-security", Action: "block"},
						},
					},
				},
			},
		},
	}

	var state SecurityProfileResourceModel
	var diags diag.Diagnostics
	mapProfileToState(ctx, profile, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if len(state.AiSecurityProfiles) != 1 {
		t.Fatalf("expected 1 ai_security_profile, got %d", len(state.AiSecurityProfiles))
	}

	asp := state.AiSecurityProfiles[0]
	assertStringValue(t, "ModelType", asp.ModelType, "default")
	assertBoolValue(t, "MaskDataInStorage", asp.MaskDataInStorage, true)

	if asp.Latency == nil {
		t.Fatal("expected latency block")
	}
	assertStringValue(t, "InlineTimeoutAction", asp.Latency.InlineTimeoutAction, "allow")
	if asp.Latency.MaxInlineLatency.ValueInt64() != 30 {
		t.Errorf("MaxInlineLatency: expected 30, got %d", asp.Latency.MaxInlineLatency.ValueInt64())
	}

	if len(asp.ModelProtection) != 2 {
		t.Fatalf("expected 2 model_protection blocks, got %d", len(asp.ModelProtection))
	}
	assertStringValue(t, "ModelProtection[0].Name", asp.ModelProtection[0].Name, "prompt-injection")
	assertStringValue(t, "ModelProtection[0].Action", asp.ModelProtection[0].Action, "block")
	assertStringValue(t, "ModelProtection[1].Name", asp.ModelProtection[1].Name, "toxic-content")
	assertStringValue(t, "ModelProtection[1].Action", asp.ModelProtection[1].Action, "high:block, moderate:allow")
	if len(asp.ModelProtection[1].ToxicCategories) != 2 {
		t.Fatalf("expected 2 toxic_category blocks, got %d", len(asp.ModelProtection[1].ToxicCategories))
	}
	assertStringValue(t, "ToxicCategory[0].Category", asp.ModelProtection[1].ToxicCategories[0].Category, "harassment")
	assertStringValue(t, "ToxicCategory[0].Action", asp.ModelProtection[1].ToxicCategories[0].Action, "block")

	if len(asp.AgentProtection) != 1 {
		t.Fatalf("expected 1 agent_protection block, got %d", len(asp.AgentProtection))
	}
	assertStringValue(t, "AgentProtection[0].Name", asp.AgentProtection[0].Name, "agent-security")
}

func TestPlanToSDKPolicy_minimal(t *testing.T) {
	ctx := context.Background()
	plan := &SecurityProfileResourceModel{}

	var diags diag.Diagnostics
	policy := planToSDKPolicy(ctx, plan, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if policy != nil {
		t.Error("expected nil policy for empty plan")
	}
}

func TestPlanToSDKPolicy_withModelProtection(t *testing.T) {
	ctx := context.Background()
	plan := &SecurityProfileResourceModel{
		AiSecurityProfiles: []AiSecurityProfileModel{
			{
				ModelType: types.StringValue("default"),
				Latency: &LatencyModel{
					InlineTimeoutAction: types.StringValue("allow"),
					MaxInlineLatency:    types.Int64Value(30),
				},
				ModelProtection: []ModelProtectionModel{
					{
						Name:   types.StringValue("prompt-injection"),
						Action: types.StringValue("block"),
					},
					{
						Name:   types.StringValue("toxic-content"),
						Action: types.StringValue("high:block, moderate:allow"),
						ToxicCategories: []ToxicCategoryModel{
							{Category: types.StringValue("harassment"), Action: types.StringValue("block")},
						},
					},
				},
			},
		},
	}

	var diags diag.Diagnostics
	policy := planToSDKPolicy(ctx, plan, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if policy == nil {
		t.Fatal("expected non-nil policy")
	}
	if len(policy.AiSecurityProfiles) != 1 {
		t.Fatalf("expected 1 ai-security-profile, got %d", len(policy.AiSecurityProfiles))
	}

	asp := policy.AiSecurityProfiles[0]
	if asp.ModelType != "default" {
		t.Errorf("ModelType: expected 'default', got %q", asp.ModelType)
	}
	if asp.ModelConfiguration == nil {
		t.Fatal("expected non-nil ModelConfiguration")
	}
	if asp.ModelConfiguration.Latency == nil {
		t.Fatal("expected non-nil Latency")
	}
	if asp.ModelConfiguration.Latency.MaxInlineLatency != 30 {
		t.Errorf("MaxInlineLatency: expected 30, got %d", asp.ModelConfiguration.Latency.MaxInlineLatency)
	}
	if len(asp.ModelConfiguration.ModelProtection) != 2 {
		t.Fatalf("expected 2 model protections, got %d", len(asp.ModelConfiguration.ModelProtection))
	}
	if asp.ModelConfiguration.ModelProtection[0].Name != "prompt-injection" {
		t.Errorf("expected 'prompt-injection', got %q", asp.ModelConfiguration.ModelProtection[0].Name)
	}
	if len(asp.ModelConfiguration.ModelProtection[1].ToxicCategoryList) != 1 {
		t.Fatalf("expected 1 toxic category, got %d", len(asp.ModelConfiguration.ModelProtection[1].ToxicCategoryList))
	}
}

func TestPlanToSDKPolicy_roundTrip(t *testing.T) {
	ctx := context.Background()
	plan := &SecurityProfileResourceModel{
		AiSecurityProfiles: []AiSecurityProfileModel{
			{
				ModelType:         types.StringValue("default"),
				MaskDataInStorage: types.BoolValue(true),
				Latency: &LatencyModel{
					InlineTimeoutAction: types.StringValue("allow"),
					MaxInlineLatency:    types.Int64Value(15),
				},
				ModelProtection: []ModelProtectionModel{
					{
						Name:   types.StringValue("prompt-injection"),
						Action: types.StringValue("block"),
					},
				},
				AgentProtection: []AgentProtectionModel{
					{Name: types.StringValue("agent-security"), Action: types.StringValue("block")},
				},
			},
		},
	}

	var diags diag.Diagnostics
	sdkPolicy := planToSDKPolicy(ctx, plan, &diags)
	if diags.HasError() {
		t.Fatalf("planToSDK diagnostics: %v", diags)
	}

	profile := &airsruntime.SecurityProfile{
		ProfileID:      "rt-001",
		ProfileName:    "round-trip",
		Active:         true,
		Policy:         sdkPolicy,
		LastModifiedTs: "2026-03-01T00:00:00Z",
	}

	var state SecurityProfileResourceModel
	mapProfileToState(ctx, profile, &state, &diags)
	if diags.HasError() {
		t.Fatalf("mapToState diagnostics: %v", diags)
	}

	if len(state.AiSecurityProfiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(state.AiSecurityProfiles))
	}
	asp := state.AiSecurityProfiles[0]
	assertStringValue(t, "ModelType", asp.ModelType, "default")
	assertBoolValue(t, "MaskDataInStorage", asp.MaskDataInStorage, true)
	if asp.Latency == nil {
		t.Fatal("expected latency")
	}
	if asp.Latency.MaxInlineLatency.ValueInt64() != 15 {
		t.Errorf("MaxInlineLatency: expected 15, got %d", asp.Latency.MaxInlineLatency.ValueInt64())
	}
	if len(asp.ModelProtection) != 1 {
		t.Fatalf("expected 1 model protection, got %d", len(asp.ModelProtection))
	}
	if len(asp.AgentProtection) != 1 {
		t.Fatalf("expected 1 agent protection, got %d", len(asp.AgentProtection))
	}
}

// mapTopicToState

func TestMapTopicToState_basic(t *testing.T) {
	ctx := context.Background()
	topic := &airsruntime.CustomTopic{
		TopicID:        "topic-123",
		TopicName:      "test-topic",
		Description:    "Test description",
		Examples:       []string{"example1", "example2"},
		CreatedTs:      "2026-01-01T00:00:00Z",
		LastModifiedTs: "2026-01-02T00:00:00Z",
	}

	var state CustomTopicResourceModel
	var diags diag.Diagnostics
	mapTopicToState(ctx, topic, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	assertStringValue(t, "ID", state.ID, "topic-123")
	assertStringValue(t, "TopicID", state.TopicID, "topic-123")
	assertStringValue(t, "TopicName", state.TopicName, "test-topic")
	assertStringValue(t, "Description", state.Description, "Test description")
	assertStringValue(t, "CreatedAt", state.CreatedAt, "2026-01-01T00:00:00Z")
	assertStringValue(t, "UpdatedAt", state.UpdatedAt, "2026-01-02T00:00:00Z")

	if state.Examples.IsNull() {
		t.Error("Examples: expected non-null list")
	}
	elems := state.Examples.Elements()
	if len(elems) != 2 {
		t.Fatalf("Examples: expected 2 elements, got %d", len(elems))
	}
}

func TestMapTopicToState_noExamples(t *testing.T) {
	ctx := context.Background()
	topic := &airsruntime.CustomTopic{
		TopicID:  "topic-456",
		Examples: nil,
	}

	var state CustomTopicResourceModel
	var diags diag.Diagnostics
	mapTopicToState(ctx, topic, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !state.Examples.IsNull() {
		t.Error("Examples: expected null list for nil examples")
	}
}

func TestMapTopicToState_emptyExamples(t *testing.T) {
	ctx := context.Background()
	topic := &airsruntime.CustomTopic{
		TopicID:  "topic-789",
		Examples: []string{},
	}

	var state CustomTopicResourceModel
	var diags diag.Diagnostics
	mapTopicToState(ctx, topic, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// empty slice → null list
	if !state.Examples.IsNull() {
		t.Error("Examples: expected null list for empty examples")
	}
}

// mapSecurityGroupToState

func assertStringValue(t *testing.T, field string, got types.String, want string) {
	t.Helper()
	if got.ValueString() != want {
		t.Errorf("%s: expected %q, got %q", field, want, got.ValueString())
	}
}

func assertBoolValue(t *testing.T, field string, got types.Bool, want bool) {
	t.Helper()
	if got.ValueBool() != want {
		t.Errorf("%s: expected %v, got %v", field, want, got.ValueBool())
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
