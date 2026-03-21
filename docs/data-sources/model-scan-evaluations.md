# prisma-airs_model_scan_evaluations

Reads evaluation results for a model security scan.

## Example Usage

```hcl
data "prisma-airs_model_scan_evaluations" "results" {
  scan_uuid = prisma-airs_model_scan.scan_bert.id
}

output "passed_count" {
  value = length([for e in data.prisma-airs_model_scan_evaluations.results.items : e if e.result == "PASSED"])
}
```

## Argument Reference

- `scan_uuid` - (Required) UUID of the scan to get evaluations for.
- `limit` - (Optional) Maximum number of evaluations to return.
- `skip` - (Optional) Number of evaluations to skip.
- `result` - (Optional) Filter by result (`PASSED`, `FAILED`, `ERROR`).

## Attribute Reference

- `items` - List of evaluations. Each item contains:
    - `uuid` - Evaluation UUID.
    - `scan_uuid` - Scan UUID.
    - `rule_instance_uuid` - Rule instance UUID.
    - `rule_name` - Rule name.
    - `result` - Evaluation result.
    - `details` - Evaluation details (JSON).
    - `created_at` - Timestamp.
- `total` - Total number of evaluations.
