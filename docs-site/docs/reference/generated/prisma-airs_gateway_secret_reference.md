# prisma-airs_gateway_secret_reference schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages an external secret reference, not the upstream secret itself. Sensitive desired auth is retained through masked reads and cannot be recovered by import.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `allow_all_workspaces` | `bool` | optional, computed | — | Allow every workspace; false is explicit. |
| `allowed_workspaces` | `set(string)` | optional, computed | — | Explicitly allowed workspace UUIDs; an empty list removes all. |
| `auth_config` | `dynamic` | required | yes | Native HCL external manager authentication settings. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `manager_type` | `string` | required | — | Secret manager type. |
| `name` | `string` | required | — | Reference name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `secret_key` | `string` | optional, computed | — | Optional key within the upstream secret. |
| `secret_path` | `string` | required | — | External secret path. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `tags` | `dynamic` | optional, computed | — | Native HCL string tag map. |
