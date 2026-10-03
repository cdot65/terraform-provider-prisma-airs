# prisma-airs_gateway_workspaces schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Lists safe admin-plane workspace metadata. The captured route has no paging parameters; incomplete results cannot prove absence.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `complete` | `bool` | computed | — | True only when the response explicitly establishes a complete inventory. |
| `has_more` | `bool` | computed | — | Whether more records are reported; null if the response omits this flag. |
| `items` | `list(object)` | computed | — | Safe metadata for the returned workspace page. |
| `status` | `string` | optional | — | Lifecycle filter; defaults to active. |
| `total_count` | `number` | computed | — | Reported inventory total. |

### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Remote workspace created_at when available. |
| `description` | `string` | computed | — | Remote workspace description when available. |
| `id` | `string` | computed | — | Remote workspace id when available. |
| `is_default` | `bool` | computed | — | Whether this is the tenant default workspace. |
| `last_updated_at` | `string` | computed | — | Remote workspace last_updated_at when available. |
| `name` | `string` | computed | — | Remote workspace name when available. |
| `scope_name` | `string` | computed | — | Remote workspace scope_name when available. |
| `slug` | `string` | computed | — | Remote workspace slug when available. |
| `status` | `string` | computed | — | Remote workspace status when available. |
