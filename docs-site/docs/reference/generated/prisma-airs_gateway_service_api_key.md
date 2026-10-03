# prisma-airs_gateway_service_api_key schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Gateway service API key through its explicit ownership route. One-time key material is sensitive and preserved through masked reads. Rotation is not automatic; deliberate replacement creates new credentials.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_emails` | `set(string)` | optional, computed | — | Usage alert recipients. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `defaults` | `dynamic` | optional, computed | — | Native HCL key defaults; config_id, allow_config_override and metadata. |
| `description` | `string` | optional, computed | — | Key description. |
| `expires_at` | `string` | optional, computed | — | RFC 3339 expiry timestamp; equivalent offsets and precision are preserved on refresh. Removing this setting does not clear a remote expiry. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `key` | `string` | computed | yes | One-time key material; export securely. Import cannot recover it. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Key name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `scopes` | `set(string)` | required | — | Gateway permissions granted to the key. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `workspace_id` | `string` | required | — | Gateway workspace UUID, including a managed gateway_workspace.id reference. |
