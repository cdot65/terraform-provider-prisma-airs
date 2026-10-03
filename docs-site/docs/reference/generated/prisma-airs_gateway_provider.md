# prisma-airs_gateway_provider schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a workspace provider backed by an organisation integration. Establish an integration_workspace_binding first. Import uses workspace_uuid/provider_uuid.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `expires_at` | `string` | optional, computed | — | RFC 3339 expiry timestamp; equivalent offsets and precision are preserved on refresh. Removing this setting does not clear a remote expiry. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `integration_id` | `string` | required | — | Bound organisation integration UUID. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Provider name. |
| `note` | `string` | optional, computed | — | Provider note; an empty string clears it. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `usage_limits` | `single(object)` | optional, computed | — | Provider usage limit settings, represented as native HCL attributes. |
| `workspace_id` | `string` | required | — | Gateway workspace UUID, including a managed gateway_workspace.id reference. |

### Attributes.usage_limits

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_threshold` | `number` | optional, computed | — | Alert threshold; zero is explicit. |
| `credit_limit` | `number` | optional, computed | — | Maximum usage credits. |
| `next_usage_reset_at` | `string` | optional, computed | — | Next reset timestamp. |
| `periodic_reset` | `string` | optional, computed | — | Reset cadence, such as monthly or weekly. |
| `periodic_reset_days` | `number` | optional, computed | — | Custom reset interval in days. |
| `type` | `string` | optional, computed | — | Usage measurement type: cost or tokens. |
