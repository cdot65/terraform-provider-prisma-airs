# AI Gateway mcp integration workspace binding

Manages `prisma-airs_gateway_mcp_integration_workspace_binding` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_mcp_integration_workspace_binding" "example" {
  integration_id = var.mcp_integration_id
  workspace_id = var.workspace_id
}
```

## Ownership and lifecycle

This owns one enabled organization MCP integration/workspace pair. Creation merges that pair; an existing enabled pair must be imported. Destroy disables only this pair. MCP servers must depend on the binding so they are removed before access is disabled. It never creates or archives a workspace.

## Large workspace inventories

The current SDK exposes no paging for the workspace-binding list. If the response reports more records than it returns, or `has_more = true`, a pair outside the page cannot be verified. Terraform retains state and blocks creation or destruction before changing access; it never infers absence from that partial list. An owned pair explicitly returned with `enabled = false` confirms absence even on a partial page.

If inventory changes after a destroy write, verification may fail after the owned pair was disabled. Verify that exact integration/workspace pair in Gateway. Once it is confirmed disabled and you intend to relinquish management, run `terraform state rm prisma-airs_gateway_mcp_integration_workspace_binding.example`. This removes only Terraform ownership. Import works when the enabled pair is visible; a pair beyond the unpaged listing remains unsupported until SDK paging is available.

## Import

```bash
terraform import prisma-airs_gateway_mcp_integration_workspace_binding.example <integration_uuid>/<workspace_uuid>
```

Import requires an existing enabled pair and an active parent integration. This resource has no secret inputs or outputs. It owns explicit workspace access only: global access settings are left unchanged, so disabling this pair does not revoke access granted by a global policy.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_mcp_integration_workspace_binding.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
