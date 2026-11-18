package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// providerSlugFixture serves one workspace provider and records every write body.
func providerSlugFixture(t *testing.T, generated string) (*gatewayResource, resource.SchemaResponse, *[]document) {
	t.Helper()
	var writes []document
	saved := document{"id": "provider-id", "workspace_id": "workspace", "integration_id": "integration", "status": "active"}
	r, _ := configFixture(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" || req.Method == "PUT" {
			var b document
			if e := json.NewDecoder(req.Body).Decode(&b); e != nil {
				t.Error(e)
			}
			writes = append(writes, b)
			for k, v := range b {
				saved[k] = v
			}
			if _, ok := saved["slug"]; !ok {
				saved["slug"] = generated
			}
		}
		_ = json.NewEncoder(w).Encode(saved)
	})
	r.definition = resourceDefinition(t, "provider")
	var s resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &s)
	return r, s, &writes
}

func createProvider(t *testing.T, r *gatewayResource, s resource.SchemaResponse, slug types.String) types.Object {
	t.Helper()
	ctx := context.Background()
	model := modelFixture(t, r, map[string]attr.Value{"id": types.StringUnknown(), "name": types.StringValue("bedrock"), "workspace_id": types.StringValue("workspace"), "integration_id": types.StringValue("integration"), "slug": slug})
	st := stateFixture(t, s, model)
	var create resource.CreateResponse
	create.State = tfsdk.State{Schema: s.Schema}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}}, &create)
	noErrors(t, create.Diagnostics)
	var current types.Object
	noErrors(t, create.State.Get(ctx, &current))
	return current
}

func TestProviderSlugIsOptionalAndReplacesOnlyWhenConfigured(t *testing.T) {
	r := &gatewayResource{definition: resourceDefinition(t, "provider")}
	a, ok := r.attributes()["slug"].(schema.StringAttribute)
	if !ok || !a.Optional || !a.Computed || a.Required {
		t.Fatal("provider slug must be optional and computed")
	}
	if len(a.PlanModifiers) != 2 {
		t.Fatal("provider slug must replace when configured and keep state when omitted")
	}
	for _, kind := range []string{"integration", "config", "mcp_server"} {
		other := &gatewayResource{definition: resourceDefinition(t, kind)}
		if b := other.attributes()["slug"].(schema.StringAttribute); b.Optional {
			t.Fatalf("%s slug must remain server-assigned", kind)
		}
	}
}

func TestProviderCreateSendsConfiguredSlug(t *testing.T) {
	r, s, writes := providerSlugFixture(t, "bedrock-us-3eaaad")
	state := createProvider(t, r, s, types.StringValue("bedrock"))
	if len(*writes) != 1 || (*writes)[0]["slug"] != "bedrock" {
		t.Fatalf("configured slug not sent on create: %v", *writes)
	}
	if state.Attributes()["slug"].(types.String).ValueString() != "bedrock" {
		t.Fatal("configured slug not persisted")
	}
}

func TestProviderCreateOmitsUnsetSlugAndStoresServerSlug(t *testing.T) {
	r, s, writes := providerSlugFixture(t, "bedrock-us-3eaaad")
	state := createProvider(t, r, s, types.StringUnknown())
	if _, sent := (*writes)[0]["slug"]; sent {
		t.Fatal("unset slug must not be sent")
	}
	if state.Attributes()["slug"].(types.String).ValueString() != "bedrock-us-3eaaad" {
		t.Fatal("server-assigned slug not stored")
	}
}

func TestProviderUpdateNeverSendsSlug(t *testing.T) {
	ctx := context.Background()
	r, s, writes := providerSlugFixture(t, "unused")
	current := createProvider(t, r, s, types.StringValue("bedrock"))
	values := current.Attributes()
	values["note"] = types.StringValue("updated")
	plan := types.ObjectValueMust(current.AttributeTypes(ctx), values)
	st := stateFixture(t, s, plan)
	var update resource.UpdateResponse
	update.State = tfsdk.State{Schema: s.Schema}
	r.Update(ctx, resource.UpdateRequest{Plan: tfsdk.Plan{Schema: s.Schema, Raw: st.Raw}, State: st}, &update)
	noErrors(t, update.Diagnostics)
	if len(*writes) != 2 {
		t.Fatalf("expected create and update writes, got %d", len(*writes))
	}
	if _, sent := (*writes)[1]["slug"]; sent {
		t.Fatal("update must not send slug")
	}
}
