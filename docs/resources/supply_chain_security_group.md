---
page_title: "prisma-airs_supply_chain_security_group (Resource)"
subcategory: "AI Supply Chain Security"
---

# prisma-airs_supply_chain_security_group Resource

Manages a model security group in Prisma AIRS Model Security API.

## Example Usage

```hcl
resource "prisma-airs_supply_chain_security_group" "ml_models" {
  name        = "production-ml-models"
  description = "Security group for production ML models"
  source_type = "HUGGING_FACE"
}
```

## Argument Reference

- `name` - (Required) Name of the security group.
- `description` - (Optional) Description of the security group.
- `source_type` - (Optional) Source type for the group. Valid values: `LOCAL`, `HUGGING_FACE`, `S3`, `GCS`, `AZURE`, `ARTIFACTORY`, `GITLAB`, `ALL`.

## Attribute Reference

- `id` - The security group UUID.
- `uuid` - The security group UUID (same as `id`).
- `state` - Current state of the group (`PENDING`, `ACTIVE`).
- `created_at` - Timestamp when the group was created.
- `updated_at` - Timestamp when the group was last updated.

## Import

```bash
terraform import prisma-airs_supply_chain_security_group.ml_models <uuid>
```

## Lifecycle

Model Security entitlement is required. Changing `source_type` plans replacement; omitted descriptions default to an empty string, and explicit empty descriptions clear the observed value. Destroy tombstones the group. Tombstoned groups are removed from state on refresh and rejected during import.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_supply_chain_security_group/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

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
