# prisma-airs_model_scan

Manages a model security scan in Prisma AIRS Model Security API.

## Example Usage

```hcl
resource "prisma-airs_model_scan" "scan_bert" {
  name                = "scan-bert-base"
  source_type         = "HUGGING_FACE"
  security_group_uuid = prisma-airs_model_security_group.ml_models.id

  source = jsonencode({
    model_name = "bert-base-uncased"
  })

  labels = {
    team        = "ml-platform"
    environment = "production"
  }
}
```

## Argument Reference

- `name` - (Required) Name of the scan.
- `source_type` - (Required) Source type. Valid values: `LOCAL`, `HUGGING_FACE`, `S3`, `GCS`, `AZURE`, `ARTIFACTORY`, `GITLAB`.
- `security_group_uuid` - (Optional) UUID of the security group to use.
- `source` - (Optional) JSON-encoded source configuration.
- `labels` - (Optional) Map of labels to attach to the scan.

## Attribute Reference

- `id` - The scan UUID.
- `uuid` - The scan UUID (same as `id`).
- `eval_outcome` - Evaluation outcome (`PENDING`, `ALLOWED`, `BLOCKED`, `ERROR`).
- `eval_summary` - Evaluation summary with passed/failed/error counts.
- `created_at` - Timestamp when the scan was created.
- `updated_at` - Timestamp when the scan was last updated.

## Import

```bash
terraform import prisma-airs_model_scan.scan_bert <uuid>
```
