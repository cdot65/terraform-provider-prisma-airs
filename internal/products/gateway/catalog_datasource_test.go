package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCatalogDataSourceReadsActiveSlugLookupAndRefreshes(t *testing.T) {
	ctx := context.Background()
	openaiID := "11111111-1111-4111-8111-111111111111"
	anthropicID := "22222222-2222-4222-8222-222222222222"
	active := true
	r, _ := configFixture(t, func(w http.ResponseWriter, q *http.Request) {
		if q.Method != http.MethodGet || q.URL.Path != aisec.GatewayProviderCatalogPath || q.URL.RawQuery != "" {
			t.Errorf("unexpected catalog request: %s %s", q.Method, q.URL)
		}
		status := "active"
		if !active {
			status = "disabled"
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"data": []any{
			map[string]any{"id": openaiID, "slug": "open-ai", "name": "OpenAI", "status": status, "key": "must-not-enter-state"},
			map[string]any{"id": anthropicID, "slug": "anthropic", "name": "Anthropic", "status": "active"},
		}}); err != nil {
			t.Fatal(err)
		}
	})
	d := &catalogDataSource{client: r.client}
	var schema datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schema)
	read := func() types.Object {
		resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, &resp)
		noErrors(t, resp.Diagnostics)
		var result types.Object
		noErrors(t, resp.State.Get(ctx, &result))
		if strings.Contains(result.String(), "must-not-enter-state") {
			t.Fatal("unexpected catalog field leaked into state")
		}
		return result
	}
	first := read().Attributes()
	ids := first["ids_by_slug"].(types.Map).Elements()
	if len(ids) != 2 || ids["open-ai"] != types.StringValue(openaiID) || ids["anthropic"] != types.StringValue(anthropicID) {
		t.Fatal("catalog UUIDs were not available by exact slug")
	}
	active = false
	second := read().Attributes()
	if len(second["items"].(types.List).Elements()) != 2 || len(second["ids_by_slug"].(types.Map).Elements()) != 1 {
		t.Fatal("refresh did not retain inactive metadata and remove it from the active lookup")
	}
}

func TestCatalogDataSourceRejectsAmbiguousAndFailedResponses(t *testing.T) {
	row := `{"id":"11111111-1111-4111-8111-111111111111","slug":"open-ai","name":"OpenAI","status":"active"}`
	for name, body := range map[string]string{
		"duplicate identity": `{"data":[` + row + `,` + row + `]}`,
		"duplicate slug":     `{"data":[` + row + `,` + strings.ReplaceAll(row, "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222") + `]}`,
		"duplicate id":       `{"data":[` + row + `,` + strings.ReplaceAll(row, "open-ai", "anthropic") + `]}`,
		"empty slug":         `{"data":[` + strings.ReplaceAll(row, "open-ai", "") + `]}`,
		"empty id":           `{"data":[` + strings.ReplaceAll(row, "11111111-1111-4111-8111-111111111111", "") + `]}`,
		"failed envelope":    `{"success":false,"data":[]}`,
		"missing data":       `{}`,
		"invalid JSON":       `{"data":`,
		"http failure":       `{"message":"not found"}`,
	} {
		t.Run(name, func(t *testing.T) {
			r, _ := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
				if name == "http failure" {
					w.WriteHeader(http.StatusNotFound)
				}
				_, _ = w.Write([]byte(body))
			})
			d := &catalogDataSource{client: r.client}
			var schema datasource.SchemaResponse
			d.Schema(context.Background(), datasource.SchemaRequest{}, &schema)
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
			d.Read(context.Background(), datasource.ReadRequest{}, &resp)
			if !resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
				t.Fatal("invalid catalog published a lookup instead of failing")
			}
		})
	}
}

func TestCatalogDataSourceAcceptsEmptyCatalog(t *testing.T) {
	r, _ := configFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	})
	d := &catalogDataSource{client: r.client}
	var schema datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &schema)
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
	d.Read(context.Background(), datasource.ReadRequest{}, &resp)
	noErrors(t, resp.Diagnostics)
	var result types.Object
	noErrors(t, resp.State.Get(context.Background(), &result))
	if len(result.Attributes()["items"].(types.List).Elements()) != 0 || len(result.Attributes()["ids_by_slug"].(types.Map).Elements()) != 0 {
		t.Fatal("empty catalog did not produce empty collections")
	}
}
