# prisma-airs_gateway_rate_limits schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Lists Gateway rate_limits metadata. This is a read-only page, not an exhaustive inventory; archived records may be returned. No secrets are stored.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `current_page` | `number` | optional | — | One-based page index (default 1). |
| `items` | `list(object)` | computed | — | Safe resource metadata for the returned page. Credentials and config documents are never included. |
| `page_size` | `number` | optional | — | Results per page (default 100). |
| `total_count` | `number` | computed | — | Server-reported total where available; otherwise the number of returned items. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. |

### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ai_provider_id` | `string` | computed | — | Remote ai_provider_id when available; otherwise null. |
| `created_at` | `string` | computed | — | Remote created_at when available; otherwise null. |
| `id` | `string` | computed | — | Remote id when available; otherwise null. |
| `integration_id` | `string` | computed | — | Remote integration_id when available; otherwise null. |
| `last_updated_at` | `string` | computed | — | Remote last_updated_at when available; otherwise null. |
| `mcp_integration_id` | `string` | computed | — | Remote mcp_integration_id when available; otherwise null. |
| `name` | `string` | computed | — | Remote name when available; otherwise null. |
| `organisation_id` | `string` | computed | — | Remote organisation_id when available; otherwise null. |
| `slug` | `string` | computed | — | Remote slug when available; otherwise null. |
| `status` | `string` | computed | — | Remote status when available; otherwise null. |
| `type` | `string` | computed | — | Remote type when available; otherwise null. |
| `user_id` | `string` | computed | — | Remote user_id when available; otherwise null. |
| `workspace_id` | `string` | computed | — | Remote workspace_id when available; otherwise null. |
