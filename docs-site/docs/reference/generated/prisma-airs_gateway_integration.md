# prisma-airs_gateway_integration schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages an organisation-level Gateway AI-provider integration. Sensitive desired settings survive masked reads. Use a separate integration_workspace_binding before creating workspace providers.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ai_provider_id` | `string` | required | — | AI provider family UUID from the Gateway catalog. |
| `configurations` | `dynamic` | optional | yes | Native HCL provider settings; masked reads cannot detect arbitrary credential/configuration drift. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Integration description; an empty string clears it. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `key` | `string` | optional | yes | Desired upstream provider credential; not recoverable on import. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Integration name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `pricing_adjustments` | `dynamic` | optional, computed | — | Native HCL pricing adjustments. |
| `secret_mappings` | `dynamic` | optional, computed | — | Native HCL secret reference mappings: field and secret_reference_id. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
