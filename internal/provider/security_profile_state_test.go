package provider_test

import (
	"context"
	"testing"

	"github.com/cdot65/prisma-airs-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSecurityProfilePreDirectionalStateRemainsReadable(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(provider.New("test")())()
	response, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: "prisma-airs_runtime_security_profile", Version: 0,
		RawState: &tfprotov6.RawState{JSON: []byte(`{"id":"old-uuid","profile_id":"old-uuid","profile_name":"legacy","revision":1,"active":true,"created_at":"original","updated_at":"original","ai_security_profile":[{"model_type":"default","content_type":null,"mask_data_in_storage":false,"latency":null,"data_protection":null,"app_protection":null,"model_protection":[{"name":"prompt-injection","action":"block","toxic_category":[],"topic_list":[]}],"agent_protection":[]}],"dlp_data_profile":[]}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, diagnostic := range response.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatal(diagnostic.Detail)
		}
	}
	schema, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	state, err := response.UpgradedState.Unmarshal(schema.ResourceSchemas["prisma-airs_runtime_security_profile"].ValueType())
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]tftypes.Value
	if err = state.As(&root); err != nil {
		t.Fatal(err)
	}
	if !root["dlp_tenant_id"].IsNull() {
		t.Fatal("new metadata changed old state")
	}
	var profiles []tftypes.Value
	if err = root["ai_security_profile"].As(&profiles); err != nil {
		t.Fatal(err)
	}
	var profile map[string]tftypes.Value
	if err = profiles[0].As(&profile); err != nil {
		t.Fatal(err)
	}
	if !profile["content_type_configurations"].IsNull() || !profile["enable_full_conversation_inspection"].IsNull() {
		t.Fatal("new settings inserted into old state")
	}
	var models []tftypes.Value
	if err = profile["model_protection"].As(&models); err != nil {
		t.Fatal(err)
	}
	var detector map[string]tftypes.Value
	if err = models[0].As(&detector); err != nil {
		t.Fatal(err)
	}
	var action string
	if err = detector["action"].As(&action); err != nil {
		t.Fatal(err)
	}
	if action != "block" || !detector["severity"].IsNull() {
		t.Fatal("legacy detector changed")
	}
}
