# prisma-airs_gateway_provider schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a workspace provider backed by an organisation integration. Establish an integration_workspace_binding first. Import uses workspace_uuid/provider_uuid.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `expires_at` | `string` | optional, computed | — | Expiry timestamp. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `integration_id` | `string` | required | — | Bound organisation integration UUID. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Provider name. |
| `note` | `string` | optional, computed | — | Provider note; an empty string clears it. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `rate_limits` | `dynamic` | optional, computed | — | Native HCL provider rate limit settings. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `usage_limits` | `dynamic` | optional, computed | — | Native HCL provider usage limit settings. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
