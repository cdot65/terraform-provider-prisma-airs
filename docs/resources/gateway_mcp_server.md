---
page_title: "prisma-airs_gateway_mcp_server (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_mcp_server Resource

Manages `prisma-airs_gateway_mcp_server` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
# MCP server: Expose the integration after its workspace binding exists.
resource "prisma-airs_gateway_mcp_server" "example" {
  name               = "Example - Gateway - Development"
  workspace_id       = var.workspace_id
  mcp_integration_id = var.mcp_integration_id
}
```

## Ownership and lifecycle

Bind the organization MCP integration to this workspace first and declare `depends_on` on the binding. Workspace or MCP integration changes require replacement. Applying Terraform does not run connectivity tests, invoke tools, terminate connections, or configure capabilities/user access.

## Import

```bash
terraform import prisma-airs_gateway_mcp_server.example <workspace_uuid>/<server_uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_mcp_server/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description; an empty string clears it. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `mcp_integration_id` | `string` | required | — | Bound organisation MCP integration UUID. |
| `name` | `string` | required | — | Server name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
