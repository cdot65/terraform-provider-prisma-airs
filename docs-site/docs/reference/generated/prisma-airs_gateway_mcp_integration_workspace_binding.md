# prisma-airs_gateway_mcp_integration_workspace_binding schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Owns one organisation integration/workspace access binding. Import uses integration_uuid/workspace_uuid. Destroy disables only this owned binding; it leaves the integration, workspace and other bindings intact.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Stable resource identifier. |
| `integration_id` | `string` | required | — | Organisation integration UUID. |
| `workspace_id` | `string` | required | — | Existing workspace UUID. |
