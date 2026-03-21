# prisma-airs_red_team_scan

Manages a red team scan (job) in Prisma AIRS Red Team API.

## Example Usage

```hcl
resource "prisma-airs_red_team_scan" "static_test" {
  name      = "security-scan-q1"
  target_id = prisma-airs_red_team_target.my_app.id
  job_type  = "STATIC"
}
```

## Argument Reference

- `name` - (Required, ForceNew) Name of the scan job.
- `target_id` - (Required, ForceNew) UUID of the target to scan.
- `job_type` - (Required, ForceNew) Type of scan. Valid values: `STATIC`, `DYNAMIC`, `CUSTOM`.

## Attribute Reference

- `id` - The job UUID.
- `job_id` - The job UUID (same as `id`).
- `status` - Job status (`INIT`, `QUEUED`, `RUNNING`, `COMPLETED`, `PARTIALLY_COMPLETE`, `FAILED`, `ABORTED`).
- `total` - Total number of attack prompts.
- `completed` - Number of completed attack prompts.
- `score` - Security score.
- `asr` - Attack success rate.
- `created_at` - Timestamp when the scan was created.
- `updated_at` - Timestamp when the scan was last updated.
- `finished_at` - Timestamp when the scan finished.

## Import

```bash
terraform import prisma-airs_red_team_scan.static_test <job_id>
```
