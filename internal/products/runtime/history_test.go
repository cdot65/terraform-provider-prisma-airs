package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func mockRuntime(t *testing.T, handler http.HandlerFunc) *airsruntime.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"fixture-token","token_type":"Bearer","expires_in":3600}`))
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	client, err := airsruntime.NewClient(airsruntime.Opts{ClientID: "fixture-client", ClientSecret: "fixture-secret", TsgID: "fixture-tsg", APIEndpoint: server.URL, TokenEndpoint: server.URL + "/token"})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
func TestLatestProfileHistoryPagination(t *testing.T) {
	pages := 0
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("latest") != "false" {
			t.Error("history filter must be explicit")
		}
		pages++
		if r.URL.Query().Get("offset") == "0" {
			_ = json.NewEncoder(w).Encode(airsruntime.SecurityProfileListResponse{Items: []airsruntime.SecurityProfile{{ProfileName: "owned", ProfileID: "uuid-v2", Revision: 2, Active: true}, {ProfileName: "other", ProfileID: "other", Revision: 99}}, NextOffset: 7})
			return
		}
		if r.URL.Query().Get("offset") != "7" {
			t.Error("ignored server pagination cursor")
		}
		_ = json.NewEncoder(w).Encode(airsruntime.SecurityProfileListResponse{Items: []airsruntime.SecurityProfile{{ProfileName: "owned", ProfileID: "uuid-v10", Revision: 10, Active: false}}})
	})
	latest, err := latestNamedProfile(context.Background(), client, "owned")
	if err != nil {
		t.Fatal(err)
	}
	if latest == nil || latest.ProfileID != "uuid-v10" || latest.Active || pages != 2 {
		t.Fatalf("latest selection or pagination incorrect: pages=%d", pages)
	}
}
func TestListFailuresDoNotMeanMissing(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	})
	ctx := context.Background()
	if _, err := latestNamedProfile(ctx, client, "owned"); err == nil {
		t.Error("profile list failure was hidden")
	}
	if _, err := findCustomerAppByName(ctx, client, "owned"); err == nil {
		t.Error("app list failure was hidden")
	}
	if _, err := findApiKeyByID(ctx, client, "owned"); err == nil {
		t.Error("key list failure was hidden")
	}
	if _, err := findTopicByID(ctx, client, "owned"); err == nil {
		t.Error("topic list failure was hidden")
	}
}
func TestCustomerAppUsesListAndCursor(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/mgmt/customerapps" {
			t.Errorf("unsupported single app route: %s", r.URL.Path)
		}
		if r.URL.Query().Get("offset") == "0" {
			_ = json.NewEncoder(w).Encode(airsruntime.CustomerAppListResponse{NextOffset: 4})
			return
		}
		if r.URL.Query().Get("offset") != "4" {
			t.Error("wrong cursor")
		}
		_ = json.NewEncoder(w).Encode(airsruntime.CustomerAppListResponse{Items: []airsruntime.CustomerApp{{AppName: "owned", CustomerAppID: "uuid-app"}}})
	})
	app, err := findCustomerAppByName(context.Background(), client, "owned")
	if err != nil {
		t.Fatal(err)
	}
	if app == nil || app.CustomerAppID != "uuid-app" {
		t.Error("lookup failed")
	}
}

func TestProfilePaginationRejectsEmptyAdvancingCursor(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(airsruntime.SecurityProfileListResponse{NextOffset: 10})
	})
	if _, err := namedProfileRevisions(context.Background(), client, "owned"); err == nil {
		t.Error("empty advancing cursor must fail, not erase ownership")
	}
}
func TestProfilePaginationRejectsRepeatedPage(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(airsruntime.SecurityProfileListResponse{Items: []airsruntime.SecurityProfile{{ProfileName: "owned", ProfileID: "uuid-v1", Revision: 1}}, NextOffset: 20})
	})
	if _, err := namedProfileRevisions(context.Background(), client, "owned"); err == nil {
		t.Error("repeated page must fail")
	}
}
func TestProfileConflictRequiresExplicitImport(t *testing.T) {
	var diagnostics diag.Diagnostics
	profileOperationError("create", "owned", aisec.NewHTTPError("conflict", aisec.ClientSideError, 409), &diagnostics)
	if !diagnostics.HasError() || !strings.Contains(diagnostics.Errors()[0].Detail(), "explicitly import") {
		t.Error("conflict lacks explicit import instruction")
	}
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(airsruntime.SecurityProfileListResponse{Items: []airsruntime.SecurityProfile{{ProfileName: "owned", ProfileID: "uuid-v1", Revision: 1}}})
	})
	diagnostics = nil
	if profileNameAvailable(context.Background(), client, "owned", &diagnostics) || !diagnostics.HasError() {
		t.Error("occupied name must not be adopted")
	}
}
func TestProfileReceiptsAndReadFailureKeepIdentity(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"message":"unavailable"}`))
	})
	for _, receipt := range []*airsruntime.SecurityProfile{nil, {}, {ProfileID: "created-uuid", Revision: 2, ProfileName: "owned"}} {
		var diagnostics diag.Diagnostics
		plan := SecurityProfileResourceModel{ProfileName: types.StringValue("owned")}
		if readCreatedProfile(context.Background(), client, receipt, &plan, &diagnostics) != nil || !diagnostics.HasError() {
			t.Error("incomplete read-back must fail")
		}
		if receipt != nil && receipt.ProfileID != "" && plan.ID.ValueString() != receipt.ProfileID {
			t.Error("successful remote ownership was discarded")
		}
	}
}
func TestProfileMissingReceiptRevision(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(airsruntime.SecurityProfileListResponse{})
	})
	var diagnostics diag.Diagnostics
	plan := SecurityProfileResourceModel{ProfileName: types.StringValue("owned")}
	result := readCreatedProfile(context.Background(), client, &airsruntime.SecurityProfile{ProfileID: "created-uuid", Revision: 1}, &plan, &diagnostics)
	if result != nil || !diagnostics.HasError() || plan.ID.ValueString() != "created-uuid" {
		t.Error("receipt UUID must survive an incomplete successful list")
	}
}

func TestTopicLookupFollowsServerCursor(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "0" {
			_ = json.NewEncoder(w).Encode(airsruntime.CustomTopicListResponse{Items: []airsruntime.CustomTopic{{TopicID: "other"}}, NextOffset: 7})
			return
		}
		if r.URL.Query().Get("offset") != "7" {
			t.Error("topic cursor ignored")
		}
		_ = json.NewEncoder(w).Encode(airsruntime.CustomTopicListResponse{Items: []airsruntime.CustomTopic{{TopicID: "owned"}}})
	})
	topic, err := findTopicByID(context.Background(), client, "owned")
	if err != nil || topic == nil || topic.TopicID != "owned" {
		t.Fatalf("cursor lookup failed: %v", err)
	}
}

func TestKeyLookupUsesServerCursor(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		page := airsruntime.ApiKeyListResponse{Items: []airsruntime.ApiKey{{ApiKeyID: "other"}}, NextOffset: 7}
		if r.URL.Query().Get("offset") == "7" {
			page = airsruntime.ApiKeyListResponse{Items: []airsruntime.ApiKey{{ApiKeyID: "owned"}}}
		}
		_ = json.NewEncoder(w).Encode(page)
	})
	key, err := findApiKeyByID(context.Background(), client, "owned")
	if err != nil || key == nil || key.ApiKeyID != "owned" {
		t.Fatalf("key cursor lookup failed: %v", err)
	}
}

func TestCustomerAppList404RemainsAnError(t *testing.T) {
	client := mockRuntime(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"message":"route missing"}`))
	})
	app, err := findCustomerAppByName(context.Background(), client, "owned")
	if err == nil || app != nil {
		t.Error("routing error treated as absent app")
	}
}
