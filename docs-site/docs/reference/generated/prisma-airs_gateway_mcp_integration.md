# prisma-airs_gateway_mcp_integration schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages an organisation MCP integration. Sensitive native configuration is desired input, never replaced by masked reads. Workspace access uses a separate mcp_integration_workspace_binding.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `auth_type` | `string` | required | — | MCP authentication type. |
| `configurations` | `dynamic` | optional | yes | Native HCL desired connection settings; arbitrary masked configuration drift cannot be detected. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description; an empty string clears it. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | MCP integration name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `secret_mappings` | `dynamic` | optional, computed | — | Native HCL secret reference mappings. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `transport` | `string` | required | — | MCP transport type. |
| `url` | `string` | required | — | MCP endpoint URL. |
