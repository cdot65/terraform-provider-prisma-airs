# prisma-airs_supply_chain_security_group schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Model Security group.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description of the security group. |
| `id` | `string` | computed | — | Terraform resource ID (same as uuid). |
| `name` | `string` | required | — | Name of the security group. |
| `source_type` | `string` | optional, computed | — | Source type (LOCAL, HUGGING_FACE, S3, GCS, AZURE, ARTIFACTORY, GITLAB, ALL). |
| `state` | `string` | computed | — | Group state (PENDING, ACTIVE). |
| `updated_at` | `string` | computed | — | Last update timestamp. |
| `uuid` | `string` | computed | — | The unique identifier of the security group. |
