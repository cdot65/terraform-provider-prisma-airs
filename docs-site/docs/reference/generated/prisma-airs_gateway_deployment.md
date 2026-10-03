# prisma-airs_gateway_deployment schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Gateway deployment registration. Destroy archives and verifies lifecycle status. It does not provision infrastructure, connect it, rotate auth or change an existing default deployment automatically.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `auth_settings` | `dynamic` | optional | yes | Native HCL desired inbound authentication settings. |
| `client_auth` | `string` | computed | yes | One-time client auth; import cannot recover it. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `credentials` | `dynamic` | computed | yes | One-time deployment credentials; import cannot recover them. |
| `deployment_config` | `dynamic` | optional | yes | Native HCL desired deployment settings; retained through masked reads. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `is_default` | `bool` | required | — | Whether to set this as default; use false for disposable registrations. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Registration name. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `tags` | `dynamic` | optional, computed | — | Native HCL string tag map. |
| `type` | `string` | required | — | Deployment type, such as non_production. |
