# prisma-airs_model_security_group

Manages a model security group in Prisma AIRS Model Security API.

## Example Usage

```hcl
resource "prisma-airs_model_security_group" "ml_models" {
  name        = "production-ml-models"
  description = "Security group for production ML model scanning"
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
terraform import prisma-airs_model_security_group.ml_models <uuid>
```
