# prisma-airs_red_team_scan

Manages a red team scan (job) in Prisma AIRS Red Team API.

## Example Usage

### Static Scan

```hcl
resource "prisma-airs_red_team_scan" "static_test" {
  target_uuid = prisma-airs_red_team_target.my_app.id
  job_type    = "STATIC"

  categories = ["SECURITY", "SAFETY"]
}
```

### Dynamic Scan

```hcl
resource "prisma-airs_red_team_scan" "dynamic_test" {
  target_uuid = prisma-airs_red_team_target.my_app.id
  job_type    = "DYNAMIC"

  categories = ["SECURITY", "SAFETY", "COMPLIANCE"]
}
```

### Custom Scan

```hcl
resource "prisma-airs_red_team_scan" "custom_test" {
  target_uuid    = prisma-airs_red_team_target.my_app.id
  job_type       = "CUSTOM"
  prompt_set_uuid = prisma-airs_red_team_custom_prompt_set.my_prompts.id
}
```

## Argument Reference

- `target_uuid` - (Required) UUID of the target to scan.
- `job_type` - (Required) Type of scan. Valid values: `STATIC`, `DYNAMIC`, `CUSTOM`.
- `categories` - (Optional) List of attack categories. Valid values: `SECURITY`, `SAFETY`, `COMPLIANCE`, `BRAND`.
- `prompt_set_uuid` - (Optional) UUID of custom prompt set (required for `CUSTOM` job type).

## Attribute Reference

- `id` - The job ID.
- `status` - Job status (`INIT`, `QUEUED`, `RUNNING`, `COMPLETED`, `PARTIALLY_COMPLETE`, `FAILED`, `ABORTED`).
- `created_at` - Timestamp when the scan was created.
- `updated_at` - Timestamp when the scan was last updated.

## Import

```bash
terraform import prisma-airs_red_team_scan.static_test <job_id>
```
