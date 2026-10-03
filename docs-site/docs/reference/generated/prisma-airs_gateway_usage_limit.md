# prisma-airs_gateway_usage_limit schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Gateway usage policy. Type and grouping are immutable. Applying configuration does not reset live usage counters.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_threshold` | `number` | optional, computed | — | Usage alert threshold; zero is explicit. |
| `conditions` | `dynamic` | required | — | Native HCL condition objects with key, value and optional excludes. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `credit_limit` | `number` | required | — | Usage credit limit. |
| `group_by` | `dynamic` | required | — | Native HCL grouping objects with key. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Policy name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `periodic_reset` | `string` | optional, computed | — | Automatic reset cadence. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `type` | `string` | required | — | Usage measurement type. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
