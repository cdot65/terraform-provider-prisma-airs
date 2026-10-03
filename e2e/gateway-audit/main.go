// gateway-audit is read-only. It fails when an active disposable fixture remains
// or a listing cannot establish a complete inventory; it never prints secrets.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	s "github.com/cdot65/prisma-airs-go/aisec/gateway/schema"
)

type item struct{ ID, Name, Status string }
type collection struct {
	name  string
	paged bool
	read  func(context.Context, int64) (any, error)
}

func ptr[T any](v T) *T { return &v }
func audit(ctx context.Context, c collection, prefix string) error {
	seen := map[string]bool{}
	active, archived := 0, 0
	for page := int64(1); ; page++ {
		response, e := c.read(ctx, page)
		if e != nil {
			return fmt.Errorf("%s listing failed; absence cannot be established", c.name)
		}
		b, e := json.Marshal(response)
		if e != nil {
			return e
		}
		var report struct {
			Data    []item `json:"data"`
			Total   *int64 `json:"total"`
			HasMore *bool  `json:"has_more"`
		}
		if e = json.Unmarshal(b, &report); e != nil {
			return e
		}
		for _, v := range report.Data {
			if v.ID == "" || seen[v.ID] {
				return fmt.Errorf("%s missing identity or repeated page", c.name)
			}
			seen[v.ID] = true
			if !strings.HasPrefix(v.Name, prefix) {
				continue
			}
			if strings.EqualFold(v.Status, "archived") || strings.EqualFold(v.Status, "deleted") {
				archived++
				fmt.Printf("ARCHIVE kind=%s id=%s verified=true\n", c.name, v.ID)
			} else {
				active++
			}
		}
		if !c.paged {
			if report.HasMore != nil && *report.HasMore {
				return fmt.Errorf("%s reports unavailable further pages", c.name)
			}
			break
		}
		if report.Total != nil && int64(len(seen)) >= *report.Total {
			break
		}
		if len(report.Data) < 100 {
			break
		}
	}
	fmt.Printf("AUDIT kind=%s active_matches=%d archived_matches=%d read_only=true\n", c.name, active, archived)
	if active > 0 {
		return fmt.Errorf("%s has active disposable fixtures", c.name)
	}
	return nil
}
func run(ctx context.Context, c *gw.Client, w, prefix string) error {
	collections := []collection{
		{"configs", false, func(ctx context.Context, _ int64) (any, error) {
			return c.Configs.List(ctx, s.ConfigsListOptions{WorkspaceID: w})
		}},
		{"guardrails", true, func(ctx context.Context, p int64) (any, error) {
			return c.Guardrails.List(ctx, s.GuardrailsListOptions{WorkspaceID: &w, PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"org_guardrails", true, func(ctx context.Context, p int64) (any, error) {
			return c.OrgGuardrails.List(ctx, s.OrgGuardrailsListOptions{PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"integrations", true, func(ctx context.Context, p int64) (any, error) {
			return c.Integrations.List(ctx, s.IntegrationsListOptions{PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"providers", true, func(ctx context.Context, p int64) (any, error) {
			return c.Providers.List(ctx, s.ProvidersListOptions{WorkspaceID: &w, PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"mcp_integrations", true, func(ctx context.Context, p int64) (any, error) {
			return c.MCPIntegrations.List(ctx, s.MCPIntegrationsListOptions{PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"mcp_servers", true, func(ctx context.Context, p int64) (any, error) {
			return c.MCPServers.List(ctx, s.MCPServersListOptions{WorkspaceID: &w, PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"service_keys", true, func(ctx context.Context, p int64) (any, error) {
			return c.APIKeys.ListForKind(ctx, gw.APIKeyService, s.APIKeysListOptions{WorkspaceID: &w, PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"user_keys", true, func(ctx context.Context, p int64) (any, error) {
			return c.APIKeys.ListForKind(ctx, gw.APIKeyUser, s.APIKeysListOptions{WorkspaceID: &w, PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"usage_limits", true, func(ctx context.Context, p int64) (any, error) {
			return c.UsageLimits.List(ctx, s.UsageLimitsListOptions{WorkspaceID: &w, PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"rate_limits", true, func(ctx context.Context, p int64) (any, error) {
			return c.RateLimits.List(ctx, s.RateLimitsListOptions{WorkspaceID: &w, PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"secret_references", true, func(ctx context.Context, p int64) (any, error) {
			return c.SecretReferences.List(ctx, s.SecretReferencesListOptions{PageSize: ptr(int64(100)), CurrentPage: &p})
		}},
		{"deployments", false, func(ctx context.Context, _ int64) (any, error) {
			return c.Deployments.List(ctx, s.DeploymentsListOptions{})
		}},
	}
	var failures []string
	for _, collection := range collections {
		if e := audit(ctx, collection, prefix); e != nil {
			failures = append(failures, e.Error())
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}
func main() {
	prefix := flag.String("prefix", "", "exact disposable run prefix (required)")
	workspace := flag.String("workspace", os.Getenv("PANW_AI_GW_TEST_WORKSPACE_ID"), "existing workspace UUID")
	flag.Parse()
	if *prefix == "" || *workspace == "" {
		fmt.Fprintln(os.Stderr, "a disposable prefix and existing workspace UUID are required")
		os.Exit(1)
	}
	c, e := gw.NewClient(gw.Opts{})
	if e != nil {
		fmt.Fprintln(os.Stderr, "Gateway client configuration failed")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if e = run(ctx, c, *workspace, *prefix); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
