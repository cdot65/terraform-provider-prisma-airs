# prisma-airs_gateway_config schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a versioned Gateway routing config using a native HCL object. Updates replace the whole document, retain the resource ID and change the computed version ID. Plaintext credentials are prohibited; reference integrations, providers or secret references.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `config` | `dynamic` | required | — | Complete native HCL routing document, without plaintext credentials. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Config name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `version_id` | `string` | computed | — | Current immutable config revision UUID; changes on update. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
