package gateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec"
	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	s "github.com/cdot65/prisma-airs-go/aisec/gateway/schema"
)

func required(kind, description string) field {
	return field{kind: kind, required: true, description: description}
}
func optional(kind, description string) field { return field{kind: kind, description: description} }
func immutable(kind, description string) field {
	f := required(kind, description)
	f.immutable = true
	return f
}
func secret(kind, description string) field {
	return field{kind: kind, sensitive: true, retained: true, description: description}
}
func computed(kind, description string) field {
	return field{kind: kind, computed: true, description: description}
}
func stableComputed(kind, description string) field {
	f := computed(kind, description)
	f.stable = true
	return f
}
func common(fields map[string]field, workspace bool) map[string]field {
	fields["id"] = stableComputed("string", "Stable resource identifier.")
	fields["slug"] = stableComputed("string", "Server-assigned resource slug.")
	fields["status"] = computed("string", "Remote lifecycle status; externally archived objects leave Terraform state.")
	fields["created_at"] = stableComputed("string", "Creation timestamp.")
	fields["last_updated_at"] = computed("string", "Last remote update timestamp.")
	fields["organisation_id"] = stableComputed("string", "Internal organisation UUID from reads; writes use the shared TSG ID.")
	if workspace {
		fields["workspace_id"] = immutable("string", "Existing Gateway workspace UUID. Workspace and IAM provisioning are external.")
	}
	return fields
}
func ptr[T any](v T) *T { return &v }

func definitions() []definition {
	config := definition{name: "config", description: "Manages a versioned Gateway routing config using a native HCL object. Updates replace the whole document, retain the resource ID and change the computed version ID. Plaintext credentials are prohibited; reference integrations, providers or secret references.", workspace: true,
		fields: common(map[string]field{"name": required("string", "Config name."), "config": required("object", "Complete native HCL routing document, without plaintext credentials."), "version_id": computed("string", "Current immutable config revision UUID; changes on update.")}, true),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.Configs.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.Configs.Get(ctx, id))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.ConfigsUpdateRequest) (*s.ConfigsUpdateResponse, error) {
				return c.sdk.Configs.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.Configs.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, w string, _, _ int64) ([]document, int64, error) {
			return listSDK(c.sdk.Configs.List(ctx, s.ConfigsListOptions{WorkspaceID: w}))
		},
	}
	result := []definition{config}
	for _, org := range []bool{false, true} {
		name := "guardrail"
		if org {
			name = "org_guardrail"
		}
		fields := common(map[string]field{"name": required("string", "Guardrail name."), "checks": required("checks", "Native HCL check objects with id and optional name and is_enabled."), "check_parameters": secret("object", "Native HCL desired parameter objects keyed by check ID; sensitive and retained through masked reads."), "actions": required("actions", "Native HCL action object: deny, async, on_success.feedback and on_fail.feedback."), "target": stableComputed("string", "Server-reported guardrail target."), "version_id": computed("string", "Guardrail revision UUID.")}, !org)
		d := definition{name: name, description: "Manages a Gateway " + strings.ReplaceAll(name, "_", " ") + " with native HCL checks and actions. Import by UUID. Destroy verifies remote absence.", fields: fields, workspace: !org, pagination: true}
		if org {
			d.create = func(ctx context.Context, c *client, b document) (document, error) {
				return writeSDK(ctx, b, c.sdk.OrgGuardrails.Create)
			}
			d.read = func(ctx context.Context, c *client, id, _ string) (document, error) {
				return readSDK(c.sdk.OrgGuardrails.Get(ctx, id))
			}
			d.update = func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
				return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateGuardrailRequest) (*s.UpdateGuardrailResponse, error) {
					return c.sdk.OrgGuardrails.Update(ctx, id, q)
				})
			}
			d.delete = func(ctx context.Context, c *client, id, _ string) error { return c.sdk.OrgGuardrails.Delete(ctx, id) }
			d.list = func(ctx context.Context, c *client, _ string, size, page int64) ([]document, int64, error) {
				return listSDK(c.sdk.OrgGuardrails.List(ctx, s.OrgGuardrailsListOptions{PageSize: ptr(size), CurrentPage: ptr(page)}))
			}
		} else {
			d.create = func(ctx context.Context, c *client, b document) (document, error) {
				return writeSDK(ctx, b, c.sdk.Guardrails.Create)
			}
			d.read = func(ctx context.Context, c *client, id, _ string) (document, error) {
				return readSDK(c.sdk.Guardrails.Get(ctx, id))
			}
			d.update = func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
				return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateGuardrailRequest) (*s.UpdateGuardrailResponse, error) {
					return c.sdk.Guardrails.Update(ctx, id, q)
				})
			}
			d.delete = func(ctx context.Context, c *client, id, _ string) error { return c.sdk.Guardrails.Delete(ctx, id) }
			d.list = func(ctx context.Context, c *client, w string, size, page int64) ([]document, int64, error) {
				return listSDK(c.sdk.Guardrails.List(ctx, s.GuardrailsListOptions{WorkspaceID: ptr(w), PageSize: ptr(size), CurrentPage: ptr(page)}))
			}
		}
		result = append(result, d)
	}
	result = append(result, definition{name: "integration", description: "Manages an organisation-level Gateway AI-provider integration. Sensitive desired settings survive masked reads. Use a separate integration_workspace_binding before creating workspace providers.", pagination: true,
		fields: common(map[string]field{"name": required("string", "Integration name."), "ai_provider_id": immutable("string", "AI provider family UUID from the Gateway catalog."), "key": secret("string", "Desired upstream provider credential; not recoverable on import."), "description": optional("string", "Integration description; an empty string clears it."), "configurations": secret("object", "Native HCL provider settings; masked reads cannot detect arbitrary credential/configuration drift."), "secret_mappings": optional("array", "Native HCL secret reference mappings: field and secret_reference_id."), "pricing_adjustments": optional("object", "Native HCL pricing adjustments.")}, false),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			b["create_default_provider"] = false
			return writeSDK(ctx, b, c.sdk.Integrations.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.Integrations.Get(ctx, id))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateIntegrationRequest) (*s.IntegrationsUpdateResponse, error) {
				return c.sdk.Integrations.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.Integrations.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, _ string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.Integrations.List(ctx, s.IntegrationsListOptions{PageSize: ptr(size), CurrentPage: ptr(page)}))
		},
	}, definition{name: "provider", description: "Manages a workspace provider backed by an organisation integration. Establish an integration_workspace_binding first. Import uses workspace_uuid/provider_uuid.", workspace: true, pagination: true, scopedImport: true,
		fields: common(map[string]field{"name": required("string", "Provider name."), "integration_id": immutable("string", "Bound organisation integration UUID."), "note": optional("string", "Provider note; an empty string clears it."), "usage_limits": optional("object", "Native HCL provider usage limit settings."), "rate_limits": optional("object", "Native HCL provider rate limit settings."), "expires_at": optional("string", "Expiry timestamp.")}, true),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.Providers.Create)
		},
		read: func(ctx context.Context, c *client, id, w string) (document, error) {
			return readSDK(c.sdk.Providers.Get(ctx, id, s.ProvidersGetOptions{WorkspaceID: ptr(w)}))
		},
		update: func(ctx context.Context, c *client, id, w string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.ProvidersUpdateRequest) (*s.ProvidersUpdateResponse, error) {
				return c.sdk.Providers.Update(ctx, id, q, s.ProvidersUpdateOptions{WorkspaceID: ptr(w)})
			})
		},
		delete: func(ctx context.Context, c *client, id, w string) error {
			_, e := c.sdk.Providers.Delete(ctx, id, s.ProvidersDeleteOptions{WorkspaceID: ptr(w)})
			return e
		},
		list: func(ctx context.Context, c *client, w string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.Providers.List(ctx, s.ProvidersListOptions{WorkspaceID: ptr(w), PageSize: ptr(size), CurrentPage: ptr(page)}))
		},
	}, definition{name: "mcp_integration", description: "Manages an organisation MCP integration. Sensitive native configuration is desired input, never replaced by masked reads. Workspace access uses a separate mcp_integration_workspace_binding.", pagination: true,
		fields: common(map[string]field{"name": required("string", "MCP integration name."), "description": optional("string", "Description; an empty string clears it."), "url": required("string", "MCP endpoint URL."), "auth_type": required("string", "MCP authentication type."), "transport": required("string", "MCP transport type."), "configurations": secret("object", "Native HCL desired connection settings; arbitrary masked configuration drift cannot be detected."), "secret_mappings": optional("array", "Native HCL secret reference mappings.")}, false),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.MCPIntegrations.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.MCPIntegrations.Get(ctx, id))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateMCPIntegration) (*s.MCPIntegrationsUpdateResponse, error) {
				return c.sdk.MCPIntegrations.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.MCPIntegrations.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, _ string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.MCPIntegrations.List(ctx, s.MCPIntegrationsListOptions{PageSize: ptr(size), CurrentPage: ptr(page)}))
		},
	}, definition{name: "mcp_server", description: "Manages a workspace MCP server backed by a bound organisation MCP integration. Connectivity tests and connection termination are operational and are not run automatically.", workspace: true, pagination: true, scopedImport: true,
		fields: common(map[string]field{"name": required("string", "Server name."), "description": optional("string", "Description; an empty string clears it."), "mcp_integration_id": immutable("string", "Bound organisation MCP integration UUID.")}, true),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.MCPServers.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.MCPServers.Get(ctx, id))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateMCPServer) (*s.MCPServersUpdateResponse, error) {
				return c.sdk.MCPServers.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.MCPServers.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, w string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.MCPServers.List(ctx, s.MCPServersListOptions{WorkspaceID: ptr(w), PageSize: ptr(size), CurrentPage: ptr(page)}))
		},
	})
	for _, kind := range []gw.APIKeyKind{gw.APIKeyService, gw.APIKeyUser} {
		fields := common(map[string]field{"name": required("string", "Key name."), "description": optional("string", "Key description."), "scopes": required("strings", "Gateway permissions granted to the key."), "expires_at": optional("string", "Expiry timestamp."), "alert_emails": optional("strings", "Usage alert recipients."), "defaults": optional("object", "Native HCL key defaults; config_id, allow_config_override and metadata."), "key": computed("string", "One-time key material; export securely. Import cannot recover it.")}, true)
		f := fields["key"]
		f.sensitive = true
		fields["key"] = f
		if kind == gw.APIKeyUser {
			fields["user_id"] = immutable("string", "User UUID owning this user key; required by SCM.")
		}
		result = append(result, definition{name: string(kind) + "_api_key", description: "Manages a Gateway " + string(kind) + " API key through its explicit ownership route. One-time key material is sensitive and preserved through masked reads. Rotation is not automatic; deliberate replacement creates new credentials.", fields: fields, workspace: true, pagination: true,
			create: func(ctx context.Context, c *client, b document) (document, error) {
				b["type"] = string(kind)
				return writeSDK(ctx, b, func(ctx context.Context, q s.CreateAPIKeyObject) (*s.APIKeysCreateResponse, error) {
					return c.sdk.APIKeys.Create(ctx, kind, q)
				})
			},
			read: func(ctx context.Context, c *client, id, _ string) (document, error) {
				return readSDK(c.sdk.APIKeys.GetForKind(ctx, kind, id))
			},
			update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
				return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateAPIKeyObject) (*s.APIKeysUpdateResponse, error) {
					return c.sdk.APIKeys.UpdateForKind(ctx, kind, id, q)
				})
			},
			delete: func(ctx context.Context, c *client, id, _ string) error {
				_, e := c.sdk.APIKeys.DeleteForKind(ctx, kind, id)
				return e
			},
			list: func(ctx context.Context, c *client, w string, size, page int64) ([]document, int64, error) {
				return listSDK(c.sdk.APIKeys.ListForKind(ctx, kind, s.APIKeysListOptions{WorkspaceID: ptr(w), PageSize: ptr(size), CurrentPage: ptr(page)}))
			},
		})
	}
	result = append(result, definition{name: "usage_limit", description: "Manages a Gateway usage policy. Type and grouping are immutable. Applying configuration does not reset live usage counters.", workspace: true, pagination: true,
		fields: common(map[string]field{"name": required("string", "Policy name."), "type": immutable("string", "Usage measurement type."), "conditions": required("array", "Native HCL condition objects with key, value and optional excludes."), "group_by": immutable("array", "Native HCL grouping objects with key."), "credit_limit": required("number", "Usage credit limit."), "alert_threshold": optional("number", "Usage alert threshold; zero is explicit."), "periodic_reset": optional("string", "Automatic reset cadence.")}, true),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.UsageLimits.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.UsageLimits.Get(ctx, id, s.UsageLimitsGetOptions{}))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateUsageLimitsPolicyRequest) (*s.UsageLimitsUpdateResponse, error) {
				return c.sdk.UsageLimits.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.UsageLimits.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, w string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.UsageLimits.List(ctx, s.UsageLimitsListOptions{WorkspaceID: ptr(w), PageSize: ptr(size), CurrentPage: ptr(page)}))
		},
	}, definition{name: "rate_limit", description: "Manages a Gateway rate policy without sending traffic. Type, target and grouping are immutable.", workspace: true, pagination: true,
		fields: common(map[string]field{"name": required("string", "Policy name."), "type": immutable("string", "Rate measurement type."), "unit": required("string", "Measurement unit, such as rpm."), "value": required("number", "Rate threshold; zero is explicit."), "conditions": required("array", "Native HCL condition objects."), "group_by": immutable("array", "Native HCL grouping objects."), "target": {kind: "string", immutable: true, description: "Optional policy target."}}, true),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.RateLimits.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.RateLimits.Get(ctx, id, s.RateLimitsGetOptions{}))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateRateLimitsPolicyRequest) (*s.RateLimitsUpdateResponse, error) {
				return c.sdk.RateLimits.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.RateLimits.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, w string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.RateLimits.List(ctx, s.RateLimitsListOptions{WorkspaceID: ptr(w), PageSize: ptr(size), CurrentPage: ptr(page)}))
		},
	}, definition{name: "secret_reference", description: "Manages an external secret reference, not the upstream secret itself. Sensitive desired auth is retained through masked reads and cannot be recovered by import.", pagination: true,
		fields: common(map[string]field{"name": required("string", "Reference name."), "description": optional("string", "Description."), "manager_type": immutable("string", "Secret manager type."), "auth_config": {kind: "object", required: true, sensitive: true, retained: true, description: "Native HCL external manager authentication settings."}, "secret_path": required("string", "External secret path."), "secret_key": optional("string", "Optional key within the upstream secret."), "allow_all_workspaces": optional("bool", "Allow every workspace; false is explicit."), "allowed_workspaces": optional("strings", "Explicitly allowed workspace UUIDs; an empty list removes all."), "tags": optional("object", "Native HCL string tag map.")}, false),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.SecretReferences.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.SecretReferences.Get(ctx, id))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateSecretReferenceRequest) (*s.SecretReferencesUpdateResponse, error) {
				return c.sdk.SecretReferences.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.SecretReferences.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, _ string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.SecretReferences.List(ctx, s.SecretReferencesListOptions{PageSize: ptr(size), CurrentPage: ptr(page)}))
		},
	}, definition{name: "deployment", description: "Manages a Gateway deployment registration. Destroy archives and verifies lifecycle status. It does not provision infrastructure, connect it, rotate auth or change an existing default deployment automatically.",
		fields: common(map[string]field{"name": required("string", "Registration name."), "type": required("string", "Deployment type, such as non_production."), "is_default": required("bool", "Whether to set this as default; use false for disposable registrations."), "deployment_config": secret("object", "Native HCL desired deployment settings; retained through masked reads."), "auth_settings": secret("object", "Native HCL desired inbound authentication settings."), "tags": optional("object", "Native HCL string tag map."), "client_auth": computed("string", "One-time client auth; import cannot recover it."), "credentials": computed("object", "One-time deployment credentials; import cannot recover them.")}, false),
		create: func(ctx context.Context, c *client, b document) (document, error) {
			return writeSDK(ctx, b, c.sdk.Deployments.Create)
		},
		read: func(ctx context.Context, c *client, id, _ string) (document, error) {
			return readSDK(c.sdk.Deployments.Get(ctx, id))
		},
		update: func(ctx context.Context, c *client, id, _ string, b document) (document, error) {
			return writeSDK(ctx, b, func(ctx context.Context, q s.UpdateDeploymentRequest) (*s.DeploymentsUpdateResponse, error) {
				return c.sdk.Deployments.Update(ctx, id, q)
			})
		},
		delete: func(ctx context.Context, c *client, id, _ string) error {
			_, e := c.sdk.Deployments.Delete(ctx, id)
			return e
		},
		list: func(ctx context.Context, c *client, _ string, size, page int64) ([]document, int64, error) {
			return listSDK(c.sdk.Deployments.List(ctx, s.DeploymentsListOptions{}))
		},
	})
	for i := range result {
		if result[i].name == "deployment" {
			delete(result[i].fields, "organisation_id")
			for _, k := range []string{"client_auth", "credentials"} {
				f := result[i].fields[k]
				f.sensitive = true
				result[i].fields[k] = f
			}
		}
	}
	return append(result, bindingDefinition(false), bindingDefinition(true))
}

// Bindings own one integration/workspace pair. They merge that pair rather
// than replacing another team's workspace access or creating default providers.
func bindingDefinition(mcp bool) definition {
	name := "integration_workspace_binding"
	if mcp {
		name = "mcp_" + name
	}
	read := func(ctx context.Context, c *client, id, _ string) (document, error) {
		parts := strings.Split(id, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid binding identifier")
		}
		var items []document
		var e error
		if mcp {
			if _, err := c.sdk.MCPIntegrations.Get(ctx, parts[0]); err != nil {
				return nil, err
			}
		} else {
			if _, err := c.sdk.Integrations.Get(ctx, parts[0]); err != nil {
				return nil, err
			}
		}
		if mcp {
			r, err := c.sdk.MCPIntegrations.ListWorkspaces(ctx, parts[0], s.MCPIntegrationsListWorkspacesOptions{})
			if err != nil {
				return nil, err
			}
			var d document
			d, e = decodeDocument(r)
			if e == nil {
				for _, k := range []string{"workspaces", "data"} {
					if values, ok := d[k].([]any); ok {
						for _, v := range values {
							if item, ok := v.(map[string]any); ok {
								items = append(items, item)
							}
						}
					}
				}
			}
		} else {
			items, _, e = listSDK(c.sdk.Integrations.ListWorkspaces(ctx, parts[0]))
		}
		if e != nil {
			return nil, e
		}
		for _, item := range items {
			if item["id"] == parts[1] && item["enabled"] == true {
				return document{"id": id, "integration_id": parts[0], "workspace_id": parts[1]}, nil
			}
		}
		return nil, aisec.NewHTTPError("binding absent", aisec.ClientSideError, 404)
	}
	set := func(ctx context.Context, c *client, id string, enabled bool) error {
		p := strings.Split(id, "/")
		if len(p) != 2 {
			return fmt.Errorf("invalid binding identifier")
		}
		item := map[string]any{"id": p[1], "enabled": enabled}
		b := document{"workspaces": []any{item}, "override_existing_workspace_access": false}
		if !mcp {
			b["create_default_provider"] = false
			item["create_default_provider"] = false
		}
		if mcp {
			_, e := writeSDK(ctx, b, func(ctx context.Context, q s.BulkUpdateMCPIntegrationWorkspaces) (*s.MCPIntegrationsSetWorkspacesResponse, error) {
				return c.sdk.MCPIntegrations.SetWorkspaces(ctx, p[0], q)
			})
			return e
		}
		_, e := writeSDK(ctx, b, func(ctx context.Context, q s.BulkUpdateWorkspacesRequest) (*s.IntegrationsSetWorkspacesResponse, error) {
			return c.sdk.Integrations.SetWorkspaces(ctx, p[0], q)
		})
		return e
	}
	return definition{name: name, description: "Owns one organisation integration/workspace access binding. Import uses integration_uuid/workspace_uuid. Destroy disables only this owned binding; it leaves the integration, workspace and other bindings intact.",
		fields: map[string]field{"id": stableComputed("string", "Stable resource identifier."), "integration_id": immutable("string", "Organisation integration UUID."), "workspace_id": immutable("string", "Existing workspace UUID.")},
		create: func(ctx context.Context, c *client, b document) (document, error) {
			id := b["integration_id"].(string) + "/" + b["workspace_id"].(string)
			_, e := read(ctx, c, id, "")
			if e == nil {
				return nil, aisec.NewHTTPError("binding exists; import it", aisec.ClientSideError, 409)
			}
			if !aisec.IsNotFound(e) {
				return nil, e
			}
			if e = set(ctx, c, id, true); e != nil {
				return nil, e
			}
			return document{"id": id}, nil
		},
		read: read,
		update: func(ctx context.Context, c *client, id, _ string, _ document) (document, error) {
			return read(ctx, c, id, "")
		},
		delete: func(ctx context.Context, c *client, id, _ string) error { return set(ctx, c, id, false) },
	}
}
