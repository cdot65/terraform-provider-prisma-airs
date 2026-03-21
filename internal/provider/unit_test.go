package provider

import (
	"context"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/cdot65/prisma-airs-go/aisec/modelsecurity"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// mapApiKeyToState
// ---------------------------------------------------------------------------

func TestMapApiKeyToState_basic(t *testing.T) {
	key := &management.ApiKey{
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
	key := &management.ApiKey{
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

// ---------------------------------------------------------------------------
// mapAppToState
// ---------------------------------------------------------------------------

func TestMapAppToState_basic(t *testing.T) {
	app := &management.CustomerApp{
		CustomerAppID: "app-789",
		AppName:       "test-app",
		TsgID:         "tsg-001",
		ModelName:     "gpt-4",
		CloudProvider: "aws",
		Environment:   "production",
		Status:        "active",
		CreatedBy:     "user@example.com",
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
}

func TestMapAppToState_emptyFields(t *testing.T) {
	app := &management.CustomerApp{
		CustomerAppID: "app-000",
	}

	var state CustomerAppResourceModel
	mapAppToState(app, &state)

	assertStringValue(t, "ID", state.ID, "app-000")
	assertStringValue(t, "AppName", state.AppName, "")
}

// ---------------------------------------------------------------------------
// mapProfileToState
// ---------------------------------------------------------------------------

func TestMapProfileToState_basic(t *testing.T) {
	ctx := context.Background()
	profile := &management.SecurityProfile{
		ProfileID:      "prof-123",
		ProfileName:    "default",
		Active:         true,
		Policy:         &management.ProfilePolicy{},
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
	assertStringValue(t, "CreatedAt", state.CreatedAt, "2026-01-02T00:00:00Z")
	assertStringValue(t, "UpdatedAt", state.UpdatedAt, "2026-01-02T00:00:00Z")
}

func TestMapProfileToState_nilPolicy(t *testing.T) {
	ctx := context.Background()
	profile := &management.SecurityProfile{
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
	profile := &management.SecurityProfile{
		ProfileID:      "prof-789",
		ProfileName:    "full-test",
		Active:         true,
		LastModifiedTs: "2026-03-01T00:00:00Z",
		Policy: &management.ProfilePolicy{
			AiSecurityProfiles: []management.AiSecurityProfileConfig{
				{
					ModelType: "default",
					ModelConfiguration: &management.ModelConfiguration{
						MaskDataInStorage: true,
						Latency: &management.LatencyConfig{
							InlineTimeoutAction: "allow",
							MaxInlineLatency:    30,
						},
						ModelProtection: []management.ModelProtectionConfig{
							{
								Name:   "prompt-injection",
								Action: "block",
							},
							{
								Name:   "toxic-content",
								Action: "alert",
								ToxicCategoryList: []management.ToxicCategoryConfig{
									{Category: "harassment", Action: "block"},
									{Category: "violence", Action: "alert"},
								},
							},
						},
						AgentProtection: []management.AgentProtectionConfig{
							{Name: "agent-threat", Action: "block"},
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
	if len(asp.ModelProtection[1].ToxicCategories) != 2 {
		t.Fatalf("expected 2 toxic_category blocks, got %d", len(asp.ModelProtection[1].ToxicCategories))
	}
	assertStringValue(t, "ToxicCategory[0].Category", asp.ModelProtection[1].ToxicCategories[0].Category, "harassment")
	assertStringValue(t, "ToxicCategory[0].Action", asp.ModelProtection[1].ToxicCategories[0].Action, "block")

	if len(asp.AgentProtection) != 1 {
		t.Fatalf("expected 1 agent_protection block, got %d", len(asp.AgentProtection))
	}
	assertStringValue(t, "AgentProtection[0].Name", asp.AgentProtection[0].Name, "agent-threat")
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
						Action: types.StringValue("alert"),
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
					{Name: types.StringValue("agent-threat"), Action: types.StringValue("alert")},
				},
			},
		},
	}

	var diags diag.Diagnostics
	sdkPolicy := planToSDKPolicy(ctx, plan, &diags)
	if diags.HasError() {
		t.Fatalf("planToSDK diagnostics: %v", diags)
	}

	profile := &management.SecurityProfile{
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

// ---------------------------------------------------------------------------
// mapTopicToState
// ---------------------------------------------------------------------------

func TestMapTopicToState_basic(t *testing.T) {
	ctx := context.Background()
	topic := &management.CustomTopic{
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
	topic := &management.CustomTopic{
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
	topic := &management.CustomTopic{
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

// ---------------------------------------------------------------------------
// mapSecurityGroupToState
// ---------------------------------------------------------------------------

func TestMapSecurityGroupToState_basic(t *testing.T) {
	group := &modelsecurity.ModelSecurityGroupResponse{
		UUID:        "sg-123",
		Name:        "test-group",
		Description: "Test security group",
		SourceType:  modelsecurity.SourceType("HUGGING_FACE"),
		State:       modelsecurity.ModelSecurityGroupState("ACTIVE"),
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-02T00:00:00Z",
	}

	var state ModelSecurityGroupResourceModel
	mapSecurityGroupToState(group, &state)

	assertStringValue(t, "ID", state.ID, "sg-123")
	assertStringValue(t, "UUID", state.UUID, "sg-123")
	assertStringValue(t, "Name", state.Name, "test-group")
	assertStringValue(t, "Description", state.Description, "Test security group")
	assertStringValue(t, "SourceType", state.SourceType, "HUGGING_FACE")
	assertStringValue(t, "State", state.State, "ACTIVE")
}

// ---------------------------------------------------------------------------
// mapTargetToState
// ---------------------------------------------------------------------------

func TestMapTargetToState_basic(t *testing.T) {
	target := &redteam.TargetResponse{
		UUID:             "tgt-123",
		Name:             "test-target",
		Description:      "Test target",
		TargetType:       redteam.TargetType("APPLICATION"),
		Status:           redteam.TargetStatus("ACTIVE"),
		ConnectionType:   redteam.TargetConnectionType("REST"),
		ConnectionParams: map[string]any{"url": "https://example.com"},
		CreatedAt:        "2026-01-01T00:00:00Z",
		UpdatedAt:        "2026-01-02T00:00:00Z",
	}

	var state RedTeamTargetResourceModel
	mapTargetToState(target, &state)

	assertStringValue(t, "ID", state.ID, "tgt-123")
	assertStringValue(t, "UUID", state.UUID, "tgt-123")
	assertStringValue(t, "Name", state.Name, "test-target")
	assertStringValue(t, "Description", state.Description, "Test target")
	assertStringValue(t, "TargetType", state.TargetType, "APPLICATION")
	assertStringValue(t, "Status", state.Status, "ACTIVE")
	assertStringValue(t, "ConnectionType", state.ConnectionType, "REST")
	assertStringContains(t, "ConnectionParams", state.ConnectionParams, `"url":"https://example.com"`)
}

func TestMapTargetToState_emptyOptionalFields(t *testing.T) {
	target := &redteam.TargetResponse{
		UUID:             "tgt-456",
		Name:             "minimal-target",
		TargetType:       "",
		ConnectionType:   "",
		ConnectionParams: nil,
	}

	var state RedTeamTargetResourceModel
	mapTargetToState(target, &state)

	assertStringValue(t, "ID", state.ID, "tgt-456")
	// TargetType and ConnectionType should remain zero value when empty
	if state.TargetType.ValueString() != "" && !state.TargetType.IsNull() {
		t.Errorf("TargetType: expected empty or null, got %q", state.TargetType.ValueString())
	}
	if state.ConnectionType.ValueString() != "" && !state.ConnectionType.IsNull() {
		t.Errorf("ConnectionType: expected empty or null, got %q", state.ConnectionType.ValueString())
	}
}

// ---------------------------------------------------------------------------
// mapPromptSetToState
// ---------------------------------------------------------------------------

func TestMapPromptSetToState_basic(t *testing.T) {
	ps := &redteam.CustomPromptSetResponse{
		UUID:        "ps-123",
		Name:        "test-promptset",
		Description: "Test prompt set",
		Status:      "READY",
		Active:      true,
		Archive:     false,
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-02T00:00:00Z",
	}

	var state RedTeamCustomPromptSetResourceModel
	mapPromptSetToState(ps, &state)

	assertStringValue(t, "ID", state.ID, "ps-123")
	assertStringValue(t, "UUID", state.UUID, "ps-123")
	assertStringValue(t, "Name", state.Name, "test-promptset")
	assertStringValue(t, "Description", state.Description, "Test prompt set")
	assertStringValue(t, "Status", state.Status, "READY")
	assertBoolValue(t, "Active", state.Active, true)
	assertBoolValue(t, "Archive", state.Archive, false)
}

// ---------------------------------------------------------------------------
// stringValueOrEnv
// ---------------------------------------------------------------------------

func TestStringValueOrEnv_configValue(t *testing.T) {
	val := types.StringValue("from-config")
	result := stringValueOrEnv(val, "TEST_UNUSED_ENV_VAR")
	if result != "from-config" {
		t.Errorf("expected 'from-config', got %q", result)
	}
}

func TestStringValueOrEnv_envFallback(t *testing.T) {
	t.Setenv("TEST_STRING_VALUE_OR_ENV", "from-env")
	val := types.StringNull()
	result := stringValueOrEnv(val, "TEST_STRING_VALUE_OR_ENV")
	if result != "from-env" {
		t.Errorf("expected 'from-env', got %q", result)
	}
}

func TestStringValueOrEnv_unknownFallsToEnv(t *testing.T) {
	t.Setenv("TEST_STRING_VALUE_UNKNOWN", "env-val")
	val := types.StringUnknown()
	result := stringValueOrEnv(val, "TEST_STRING_VALUE_UNKNOWN")
	if result != "env-val" {
		t.Errorf("expected 'env-val', got %q", result)
	}
}

func TestStringValueOrEnv_noConfigNoEnv(t *testing.T) {
	val := types.StringNull()
	result := stringValueOrEnv(val, "TEST_NONEXISTENT_VAR")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// Client extractor helpers
// ---------------------------------------------------------------------------

func TestGetMgmtClient_nil(t *testing.T) {
	_, diags := getMgmtClient(nil)
	if !diags.HasError() {
		t.Error("expected error for nil provider data")
	}
}

func TestGetMgmtClient_wrongType(t *testing.T) {
	_, diags := getMgmtClient("wrong-type")
	if !diags.HasError() {
		t.Error("expected error for wrong type")
	}
}

func TestGetMgmtClient_noClient(t *testing.T) {
	pd := &ProviderData{}
	_, diags := getMgmtClient(pd)
	if !diags.HasError() {
		t.Error("expected error when MgmtClient is nil")
	}
}

func TestGetMgmtClient_valid(t *testing.T) {
	pd := &ProviderData{MgmtClient: &management.Client{}}
	client, diags := getMgmtClient(pd)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestGetModelSecClient_nil(t *testing.T) {
	_, diags := getModelSecClient(nil)
	if !diags.HasError() {
		t.Error("expected error for nil provider data")
	}
}

func TestGetModelSecClient_noClient(t *testing.T) {
	pd := &ProviderData{}
	_, diags := getModelSecClient(pd)
	if !diags.HasError() {
		t.Error("expected error when ModelSecClient is nil")
	}
}

func TestGetModelSecClient_valid(t *testing.T) {
	pd := &ProviderData{ModelSecClient: &modelsecurity.Client{}}
	client, diags := getModelSecClient(pd)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestGetRedTeamClient_nil(t *testing.T) {
	_, diags := getRedTeamClient(nil)
	if !diags.HasError() {
		t.Error("expected error for nil provider data")
	}
}

func TestGetRedTeamClient_noClient(t *testing.T) {
	pd := &ProviderData{}
	_, diags := getRedTeamClient(pd)
	if !diags.HasError() {
		t.Error("expected error when RedTeamClient is nil")
	}
}

func TestGetRedTeamClient_valid(t *testing.T) {
	pd := &ProviderData{RedTeamClient: &redteam.Client{}}
	client, diags := getRedTeamClient(pd)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

// ---------------------------------------------------------------------------
// Schema validation tests
// ---------------------------------------------------------------------------

func TestAllResourceSchemas(t *testing.T) {
	resources := map[string]func() interface {
		Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	}{
		"api_key": func() interface {
			Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		} {
			return &apiKeyResource{}
		},
		"customer_app": func() interface {
			Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		} {
			return &customerAppResource{}
		},
		"security_profile": func() interface {
			Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		} {
			return &securityProfileResource{}
		},
		"custom_topic": func() interface {
			Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		} {
			return &customTopicResource{}
		},
		"model_security_group": func() interface {
			Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		} {
			return &modelSecurityGroupResource{}
		},
		"red_team_target": func() interface {
			Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		} {
			return &redTeamTargetResource{}
		},
		"red_team_custom_prompt_set": func() interface {
			Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
		} {
			return &redTeamCustomPromptSetResource{}
		},
	}

	for name, factory := range resources {
		t.Run(name, func(t *testing.T) {
			r := factory()
			var resp resource.SchemaResponse
			r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

			if resp.Schema.Description == "" {
				t.Error("schema description is empty")
			}

			attrs := resp.Schema.Attributes
			if len(attrs) == 0 {
				t.Error("schema has no attributes")
			}

			// Every resource must have an "id" attribute
			if _, ok := attrs["id"]; !ok {
				t.Error("schema missing 'id' attribute")
			}
		})
	}
}

func TestAllDataSourceSchemas(t *testing.T) {
	dataSources := map[string]func() interface {
		Schema(context.Context, datasource.SchemaRequest, *datasource.SchemaResponse)
	}{
		"dlp_profiles": func() interface {
			Schema(context.Context, datasource.SchemaRequest, *datasource.SchemaResponse)
		} {
			return &dlpProfilesDataSource{}
		},
		"deployment_profiles": func() interface {
			Schema(context.Context, datasource.SchemaRequest, *datasource.SchemaResponse)
		} {
			return &deploymentProfilesDataSource{}
		},
		"model_security_rules": func() interface {
			Schema(context.Context, datasource.SchemaRequest, *datasource.SchemaResponse)
		} {
			return &modelSecurityRulesDataSource{}
		},
	}

	for name, factory := range dataSources {
		t.Run(name, func(t *testing.T) {
			ds := factory()
			var resp datasource.SchemaResponse
			ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

			if resp.Schema.Description == "" {
				t.Error("schema description is empty")
			}

			attrs := resp.Schema.Attributes
			if len(attrs) == 0 {
				t.Error("schema has no attributes")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Resource metadata tests
// ---------------------------------------------------------------------------

func TestResourceMetadata(t *testing.T) {
	tests := map[string]struct {
		resource     resource.Resource
		expectedType string
	}{
		"api_key":                    {&apiKeyResource{}, "test_api_key"},
		"customer_app":               {&customerAppResource{}, "test_customer_app"},
		"security_profile":           {&securityProfileResource{}, "test_security_profile"},
		"custom_topic":               {&customTopicResource{}, "test_custom_topic"},
		"model_security_group":       {&modelSecurityGroupResource{}, "test_model_security_group"},
		"red_team_target":            {&redTeamTargetResource{}, "test_red_team_target"},
		"red_team_custom_prompt_set": {&redTeamCustomPromptSetResource{}, "test_red_team_custom_prompt_set"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := resource.MetadataRequest{ProviderTypeName: "test"}
			var resp resource.MetadataResponse
			tc.resource.Metadata(context.Background(), req, &resp)

			if resp.TypeName != tc.expectedType {
				t.Errorf("expected type name %q, got %q", tc.expectedType, resp.TypeName)
			}
		})
	}
}

func TestDataSourceMetadata(t *testing.T) {
	tests := map[string]struct {
		ds           datasource.DataSource
		expectedType string
	}{
		"dlp_profiles":         {&dlpProfilesDataSource{}, "test_dlp_profiles"},
		"deployment_profiles":  {&deploymentProfilesDataSource{}, "test_deployment_profiles"},
		"model_security_rules": {&modelSecurityRulesDataSource{}, "test_model_security_rules"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := datasource.MetadataRequest{ProviderTypeName: "test"}
			var resp datasource.MetadataResponse
			tc.ds.Metadata(context.Background(), req, &resp)

			if resp.TypeName != tc.expectedType {
				t.Errorf("expected type name %q, got %q", tc.expectedType, resp.TypeName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Configure with nil ProviderData tests (should not panic)
// ---------------------------------------------------------------------------

func TestResourceConfigure_nilProviderData(t *testing.T) {
	resources := []resource.Resource{
		&apiKeyResource{},
		&customerAppResource{},
		&securityProfileResource{},
		&customTopicResource{},
		&modelSecurityGroupResource{},
		&redTeamTargetResource{},
		&redTeamCustomPromptSetResource{},
	}

	for _, r := range resources {
		rc, ok := r.(resource.ResourceWithConfigure)
		if !ok {
			continue
		}
		// Should not panic with nil ProviderData
		var resp resource.ConfigureResponse
		rc.Configure(context.Background(), resource.ConfigureRequest{ProviderData: nil}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("unexpected error for nil ProviderData on %T", r)
		}
	}
}

func TestDataSourceConfigure_nilProviderData(t *testing.T) {
	dataSources := []datasource.DataSource{
		&dlpProfilesDataSource{},
		&deploymentProfilesDataSource{},
		&modelSecurityRulesDataSource{},
	}

	for _, ds := range dataSources {
		dc, ok := ds.(datasource.DataSourceWithConfigure)
		if !ok {
			continue
		}
		var resp datasource.ConfigureResponse
		dc.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("unexpected error for nil ProviderData on %T", ds)
		}
	}
}

// ---------------------------------------------------------------------------
// Provider New function
// ---------------------------------------------------------------------------

func TestNewProvider(t *testing.T) {
	p := New("test-version")()
	if p == nil {
		t.Fatal("New() returned nil provider")
	}
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestResourceConstructors(t *testing.T) {
	constructors := map[string]func() resource.Resource{
		"api_key":                    NewApiKeyResource,
		"customer_app":               NewCustomerAppResource,
		"security_profile":           NewSecurityProfileResource,
		"custom_topic":               NewCustomTopicResource,
		"model_security_group":       NewModelSecurityGroupResource,
		"red_team_target":            NewRedTeamTargetResource,
		"red_team_custom_prompt_set": NewRedTeamCustomPromptSetResource,
	}

	for name, fn := range constructors {
		t.Run(name, func(t *testing.T) {
			r := fn()
			if r == nil {
				t.Error("constructor returned nil")
			}
		})
	}
}

func TestDataSourceConstructors(t *testing.T) {
	constructors := map[string]func() datasource.DataSource{
		"dlp_profiles":         NewDlpProfilesDataSource,
		"deployment_profiles":  NewDeploymentProfilesDataSource,
		"model_security_rules": NewModelSecurityRulesDataSource,
	}

	for name, fn := range constructors {
		t.Run(name, func(t *testing.T) {
			ds := fn()
			if ds == nil {
				t.Error("constructor returned nil")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

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

func assertStringContains(t *testing.T, field string, got types.String, substr string) {
	t.Helper()
	val := got.ValueString()
	if len(val) == 0 {
		t.Errorf("%s: expected string containing %q, got empty", field, substr)
		return
	}
	if !containsSubstring(val, substr) {
		t.Errorf("%s: expected string containing %q, got %q", field, substr, val)
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
