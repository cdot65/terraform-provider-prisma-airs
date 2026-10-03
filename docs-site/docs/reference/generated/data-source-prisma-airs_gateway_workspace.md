# prisma-airs_gateway_workspace schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Reads safe workspace metadata from the tenant admin plane. Does not expose defaults, credentials, users, or security settings.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

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
| `workspace_id` | `string` | required | — | Workspace UUID to read. |
