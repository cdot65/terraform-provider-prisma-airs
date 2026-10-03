# prisma-airs_gateway_mcp_server schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a workspace MCP server backed by a bound organisation MCP integration. Connectivity tests and connection termination are operational and are not run automatically.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

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
