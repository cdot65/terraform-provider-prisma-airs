# prisma-airs_model_scan_violations

Reads violation results for a model security scan.

## Example Usage

```hcl
data "prisma-airs_model_scan_violations" "results" {
  scan_uuid = prisma-airs_model_scan.scan_bert.id
}
```

## Argument Reference

- `scan_uuid` - (Required) UUID of the scan to get violations for.
- `limit` - (Optional) Maximum number of violations to return.
- `skip` - (Optional) Number of violations to skip.

## Attribute Reference

- `items` - List of violations. Each item contains:
    - `uuid` - Violation UUID.
    - `scan_uuid` - Scan UUID.
    - `rule_name` - Rule name.
    - `details` - Violation details (JSON).
    - `created_at` - Timestamp.
- `total` - Total number of violations.
