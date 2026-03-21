# prisma-airs_model_scan

Manages a model security scan in Prisma AIRS Model Security API.

## Example Usage

```hcl
resource "prisma-airs_model_scan" "scan_bert" {
  model_uri           = "hf://bert-base-uncased"
  scan_origin         = "HUGGING_FACE"
  security_group_uuid = prisma-airs_model_security_group.ml_models.id

  labels = {
    team        = "ml-platform"
    environment = "production"
  }
}
```

## Argument Reference

- `model_uri` - (Required, ForceNew) URI of the model to scan.
- `scan_origin` - (Required, ForceNew) Scan origin. Valid values: `MODEL_SECURITY_SDK`, `HUGGING_FACE`.
- `security_group_uuid` - (Optional, ForceNew) UUID of the security group to use.
- `labels` - (Optional) Map of key-value labels for the scan.

## Attribute Reference

- `id` - The scan UUID.
- `uuid` - The scan UUID (same as `id`).
- `source_type` - Source type (`LOCAL`, `HUGGING_FACE`, `S3`, `GCS`, `AZURE`, `ARTIFACTORY`, `GITLAB`, `ALL`).
- `eval_outcome` - Evaluation outcome (`PENDING`, `ALLOWED`, `BLOCKED`, `ERROR`).
- `eval_summary` - Evaluation summary as JSON with rules_passed/rules_failed/total_rules.
- `created_at` - Timestamp when the scan was created.
- `updated_at` - Timestamp when the scan was last updated.

## Import

```bash
terraform import prisma-airs_model_scan.scan_bert <uuid>
```
