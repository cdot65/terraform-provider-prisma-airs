package provider_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cdot65/prisma-airs-go/aisec"
	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccGatewayWorkspaceManagedLifecycle(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	c := accGatewayClient(t)
	scope := "tf_workspace_live_" + acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if _, err := c.IAMScopes.Get(ctx, scope); !aisec.IsNotFound(err) {
		t.Fatal("Disposable IAM name not confirmed unused")
	}
	address := "prisma-airs_gateway_workspace.application"
	childAddress := "prisma-airs_gateway_config.application"
	report := map[string]any{"scope_name": scope, "sdk_version": aisec.Version, "started_at": time.Now().UTC().Format(time.RFC3339)}
	save := func() {
		if path := os.Getenv("PANW_AI_GW_WORKSPACE_RECEIPT"); path != "" {
			b, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				t.Fatal("receipt encoding failed")
			}
			if err = os.WriteFile(path, b, 0600); err != nil {
				t.Fatal("receipt persistence failed")
			}
		}
	}
	save()
	var workspaceID, slug, childID, policyID string
	config := func(label string, credit int, settings bool) string {
		extras := ""
		if settings {
			extras = fmt.Sprintf(`icon="test"
 defaults={metadata={owner="terraform",scope=%q}}
 usage_limits=[{type="tokens",credit_limit=%d}]
 rate_limits=[{type="requests",unit="rpm",value=60},{type="tokens",unit="tpm",value=1000}]
`, scope, credit)
		}
		return fmt.Sprintf(`provider "prisma-airs" {}
resource "prisma-airs_gateway_workspace" "application" {
 name=%q
 scope_name=%q
 scope_management="managed"
 %s
}
resource "prisma-airs_gateway_config" "application" {
 name=%q
 workspace_id=prisma-airs_gateway_workspace.application.id
 config={provider="openai",retry={attempts=1}}
}
data "prisma-airs_gateway_workspace" "application" {workspace_id=prisma-airs_gateway_workspace.application.id}
data "prisma-airs_gateway_workspaces" "all" {}
`, label, scope, extras, scope)
	}
	verify := func(step string, credit int) resource.TestCheckFunc {
		return func(st *terraform.State) error {
			row := st.RootModule().Resources[address]
			if row == nil {
				return fmt.Errorf("Workspace state missing")
			}
			id := row.Primary.ID
			if workspaceID == "" {
				workspaceID = id
				slug = row.Primary.Attributes["slug"]
				report["workspace_id"] = id
				report["slug"] = slug
			} else if id != workspaceID || row.Primary.Attributes["slug"] != slug {
				return fmt.Errorf("Workspace label update replaced UUID or slug")
			}
			child := st.RootModule().Resources[childAddress]
			if child == nil {
				return fmt.Errorf("Child config state missing")
			}
			if childID == "" {
				childID = child.Primary.ID
				report["child_id"] = childID
			} else if childID != child.Primary.ID {
				return fmt.Errorf("Workspace update replaced child config")
			}
			save()
			readCtx, stop := context.WithTimeout(context.Background(), 2*time.Minute)
			defer stop()
			w, err := c.Workspaces.Get(readCtx, id, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
			if err != nil {
				return fmt.Errorf("Workspace independent read failed")
			}
			owned, err := c.IAMScopes.Get(readCtx, scope)
			if err != nil || owned.Name != scope || len(owned.Resources) != 1 || owned.Resources[0].ResourceID != slug || owned.Resources[0].ResourceType != "workspace" {
				return fmt.Errorf("Dedicated scope binding was not independently confirmed")
			}
			if _, err = c.Configs.Get(readCtx, childID); err != nil {
				return fmt.Errorf("Child read with intended tenant credentials failed")
			}
			report[step] = map[string]any{"defaults": w.Defaults, "usage_limits": w.UsageLimits, "rate_limits": w.RateLimits, "icon": w.Icon, "name": w.Name, "scope_binding_confirmed": true, "child_read_confirmed": true}
			save()
			if credit > 0 {
				if w.UsageLimits == nil {
					return fmt.Errorf("Usage policy missing")
				}
				var policies []map[string]any
				if err = json.Unmarshal(*w.UsageLimits, &policies); err != nil || len(policies) != 1 || policies[0]["credit_limit"] != float64(credit) {
					return fmt.Errorf("Usage credit not applied")
				}
				next, _ := policies[0]["id"].(string)
				if policyID == "" {
					policyID = next
				} else if next != policyID {
					return fmt.Errorf("Credit change replaced inline policy UUID")
				}
			}
			if credit == 0 {
				d := map[string]any{}
				if w.Defaults != nil {
					d = *w.Defaults
				}
				metadata, _ := d["metadata"].(map[string]any)
				if d["config_id"] != nil || len(metadata) > 0 || w.UsageLimits != nil || w.RateLimits == nil || string(*w.RateLimits) != "[]" || (w.Icon != nil && *w.Icon != "") {
					return fmt.Errorf("Removed settings remain active")
				}
			}
			return nil
		}
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config(scope, 1000, true), Check: verify("created", 1000)},
		{Config: config(scope+" renamed", 2000, true), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)}}, Check: verify("updated", 2000)},
		{Config: config(scope+" renamed", 2000, true), PlanOnly: true},
		{ResourceName: address, ImportState: true, ImportStateIdFunc: func(*terraform.State) (string, error) { return "managed/" + workspaceID + "/" + scope, nil }, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"scope_ownership_token", "defaults", "usage_limits", "rate_limits", "icon"}},
		{Config: config(scope+" renamed", 0, false), Check: verify("cleared", 0)},
		{Config: config(scope+" renamed", 0, false), PlanOnly: true},
	}, CheckDestroy: func(st *terraform.State) error {
		if workspaceID == "" {
			if row := st.RootModule().Resources[address]; row != nil {
				workspaceID = row.Primary.ID
				report["workspace_id"] = workspaceID
			}
		}
		if childID == "" {
			if row := st.RootModule().Resources[childAddress]; row != nil {
				childID = row.Primary.ID
				report["child_id"] = childID
			}
		}
		save()
		cleanup, stop := context.WithTimeout(context.Background(), 2*time.Minute)
		defer stop()
		_, workspaceErr := c.Workspaces.Get(cleanup, workspaceID, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
		_, scopeErr := c.IAMScopes.Get(cleanup, scope)
		var childErr error
		if childID != "" {
			_, childErr = c.Configs.Get(cleanup, childID)
		} else {
			childErr = aisec.NewHTTPError("child was not created", aisec.ClientSideError, 404)
		}
		report["workspace_inactive_confirmed"] = aisec.IsNotFound(workspaceErr)
		report["scope_absence_confirmed"] = aisec.IsNotFound(scopeErr)
		report["child_absence_confirmed"] = aisec.IsNotFound(childErr)
		save()
		if !aisec.IsNotFound(workspaceErr) || !aisec.IsNotFound(scopeErr) || !aisec.IsNotFound(childErr) {
			return fmt.Errorf("Independent cleanup checks failed; retained identities recorded")
		}
		return nil
	}})
	// The workspace is archived, not hard-deleted. No role grants or membership writes.
	if strings.TrimSpace(workspaceID) == "" {
		t.Fatal("workspace lifecycle did not run")
	}
}

func TestAccGatewayWorkspaceExternalScopeIsolation(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	c := accGatewayClient(t)
	scope := "tf_ws_external_" + acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if _, err := c.IAMScopes.Get(ctx, scope); !aisec.IsNotFound(err) {
		t.Fatal("Disposable external scope name not confirmed unused")
	}
	original, err := c.IAMScopes.Create(ctx, gw.IAMScopeCreateInput{Name: scope, Description: "Disposable external scope isolation fixture"})
	if err != nil {
		t.Fatal("External test fixture scope create failed")
	}
	before, err := json.Marshal(original)
	if err != nil {
		t.Fatal("External scope snapshot failed")
	}
	addr := "prisma-airs_gateway_workspace.application"
	var id string
	report := map[string]any{"scope_name": scope, "sdk_version": aisec.Version}
	save := func() {
		if base := os.Getenv("PANW_AI_GW_WORKSPACE_RECEIPT"); base != "" {
			path := strings.TrimSuffix(base, ".json") + "-external.json"
			b, e := json.MarshalIndent(report, "", "  ")
			if e != nil || os.WriteFile(path, b, 0600) != nil {
				t.Fatal("External receipt save failed")
			}
		}
	}
	save()
	config := func(label string) string {
		return fmt.Sprintf(`provider "prisma-airs" {}
resource "prisma-airs_gateway_workspace" "application" {
 name=%q
 scope_name=%q
 scope_management="external"
}
`, label, scope)
	}
	unchanged := func() error {
		readCtx, stop := context.WithTimeout(context.Background(), time.Minute)
		defer stop()
		current, e := c.IAMScopes.Get(readCtx, scope)
		if e != nil {
			return fmt.Errorf("External scope read failed")
		}
		now, e := json.Marshal(current)
		if e != nil || string(now) != string(before) {
			return fmt.Errorf("Provider changed external scope snapshot")
		}
		return nil
	}
	t.Cleanup(func() {
		if id == "" {
			t.Error("External scope fixture retained because workspace creation outcome lacks a recorded UUID")
			return
		}
		cleanup, stop := context.WithTimeout(context.Background(), time.Minute)
		defer stop()
		_, e := c.Workspaces.Get(cleanup, id, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
		if !aisec.IsNotFound(e) {
			t.Error("External fixture scope retained until workspace inactivity is confirmed")
			return
		}
		if e = c.IAMScopes.Delete(cleanup, scope); e != nil {
			t.Error("External fixture scope cleanup failed")
			return
		}
		_, e = c.IAMScopes.Get(cleanup, scope)
		report["test_fixture_scope_absence_confirmed"] = aisec.IsNotFound(e)
		save()
		if !aisec.IsNotFound(e) {
			t.Error("External fixture scope absence unconfirmed")
		}
	})
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config(scope), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "scope_owned", "false"), resource.TestCheckResourceAttr(addr, "scope_binding_ready", "false"), func(st *terraform.State) error {
			id = st.RootModule().Resources[addr].Primary.ID
			report["workspace_id"] = id
			save()
			return unchanged()
		})},
		{Config: config(scope + " renamed"), Check: func(st *terraform.State) error {
			if st.RootModule().Resources[addr].Primary.ID != id {
				return fmt.Errorf("External workspace rename replaced UUID")
			}
			return unchanged()
		}},
		{ResourceName: addr, ImportState: true, ImportStateIdFunc: func(*terraform.State) (string, error) { return id + "/" + scope, nil }, ImportStateVerify: true},
		{Config: config(scope + " renamed"), PlanOnly: true},
	}, CheckDestroy: func(st *terraform.State) error {
		if id == "" {
			if row := st.RootModule().Resources[addr]; row != nil {
				id = row.Primary.ID
				report["workspace_id"] = id
			}
		}
		cleanup, stop := context.WithTimeout(context.Background(), time.Minute)
		defer stop()
		_, e := c.Workspaces.Get(cleanup, id, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
		report["workspace_inactive_confirmed"] = aisec.IsNotFound(e)
		if !aisec.IsNotFound(e) {
			save()
			return fmt.Errorf("External workspace archival unconfirmed")
		}
		e = unchanged()
		report["external_scope_snapshot_unchanged"] = e == nil
		save()
		return e
	}})
}

func TestAccGatewayWorkspaceArchiveReplacement(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccPreCheck(t)
	c := accGatewayClient(t)
	scope := "tf_ws_replace_" + acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, err := c.IAMScopes.Get(ctx, scope); !aisec.IsNotFound(err) {
		t.Fatal("Replacement scope name not confirmed unused")
	}
	addr := "prisma-airs_gateway_workspace.application"
	var oldID, newID, oldToken string
	report := map[string]any{"scope_name": scope, "sdk_version": aisec.Version}
	save := func() {
		if base := os.Getenv("PANW_AI_GW_WORKSPACE_RECEIPT"); base != "" {
			b, e := json.MarshalIndent(report, "", "  ")
			if e != nil || os.WriteFile(strings.TrimSuffix(base, ".json")+"-replacement.json", b, 0600) != nil {
				t.Fatal("Replacement receipt persistence failed")
			}
		}
	}
	save()
	config := fmt.Sprintf(`provider "prisma-airs" {}
resource "prisma-airs_gateway_workspace" "application" {
 name=%q
 scope_name=%q
 scope_management="managed"
}
`, scope, scope)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, ExpectNonEmptyPlan: true, Check: func(st *terraform.State) error {
			row := st.RootModule().Resources[addr]
			oldID = row.Primary.ID
			oldToken = row.Primary.Attributes["scope_ownership_token"]
			report["old_workspace_id"] = oldID
			report["old_token"] = oldToken
			save()
			readCtx, stop := context.WithTimeout(context.Background(), time.Minute)
			defer stop()
			if e := c.Workspaces.Delete(readCtx, oldID); e != nil {
				return fmt.Errorf("Disposable remote archive failed")
			}
			_, e := c.Workspaces.Get(readCtx, oldID, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
			report["remote_archive_confirmed"] = aisec.IsNotFound(e)
			save()
			if !aisec.IsNotFound(e) {
				return fmt.Errorf("Remote archival unconfirmed")
			}
			return nil
		}},
		{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate)}}, Check: func(st *terraform.State) error {
			row := st.RootModule().Resources[addr]
			newID = row.Primary.ID
			token := row.Primary.Attributes["scope_ownership_token"]
			slug := row.Primary.Attributes["slug"]
			report["new_workspace_id"] = newID
			report["new_token"] = token
			report["new_slug"] = slug
			save()
			if oldID == newID || token == oldToken || token == "" {
				return fmt.Errorf("Replacement did not create new workspace/scope identities")
			}
			readCtx, stop := context.WithTimeout(context.Background(), time.Minute)
			defer stop()
			scopeRow, e := c.IAMScopes.Get(readCtx, scope)
			if e != nil || scopeRow.Description != "Terraform workspace ownership "+token || len(scopeRow.Resources) != 1 || scopeRow.Resources[0].ResourceID != slug {
				return fmt.Errorf("Replacement scope/binding identity not confirmed")
			}
			_, e = c.Workspaces.Get(readCtx, oldID, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
			report["old_workspace_inactive_confirmed"] = aisec.IsNotFound(e)
			report["new_scope_binding_confirmed"] = true
			save()
			if !aisec.IsNotFound(e) {
				return fmt.Errorf("Old workspace remains active")
			}
			return nil
		}},
		{Config: config, PlanOnly: true},
	}, CheckDestroy: func(st *terraform.State) error {
		if newID == "" {
			if row := st.RootModule().Resources[addr]; row != nil {
				newID = row.Primary.ID
				report["new_workspace_id"] = newID
			}
		}
		cleanup, stop := context.WithTimeout(context.Background(), time.Minute)
		defer stop()
		_, e1 := c.Workspaces.Get(cleanup, oldID, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
		_, e2 := c.Workspaces.Get(cleanup, newID, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin})
		_, e3 := c.IAMScopes.Get(cleanup, scope)
		report["old_workspace_inactive_confirmed"] = aisec.IsNotFound(e1)
		report["new_workspace_inactive_confirmed"] = aisec.IsNotFound(e2)
		report["scope_absence_confirmed"] = aisec.IsNotFound(e3)
		save()
		if !aisec.IsNotFound(e1) || !aisec.IsNotFound(e2) || !aisec.IsNotFound(e3) {
			return fmt.Errorf("Replacement cleanup unconfirmed")
		}
		return nil
	}})
}
