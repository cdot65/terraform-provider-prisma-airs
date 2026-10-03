# prisma-airs_gateway_rate_limit schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Gateway rate policy without sending traffic. Type, target and grouping are immutable.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `conditions` | `dynamic` | required | — | Native HCL condition objects. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `group_by` | `dynamic` | required | — | Native HCL grouping objects. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Policy name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `target` | `string` | optional, computed | — | Optional policy target. |
| `type` | `string` | required | — | Rate measurement type. |
| `unit` | `string` | required | — | Measurement unit, such as rpm. |
| `value` | `number` | required | — | Rate threshold; zero is explicit. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
