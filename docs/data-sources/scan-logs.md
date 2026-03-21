# prisma-airs_scan_logs

Reads scan activity logs from Prisma AIRS Management API.

## Example Usage

```hcl
data "prisma-airs_scan_logs" "recent" {
  limit = 50
}

output "log_count" {
  value = data.prisma-airs_scan_logs.recent.total_count
}
```

## Argument Reference

- `limit` - (Optional) Maximum number of logs to return.
- `offset` - (Optional) Offset for pagination.

## Attribute Reference

- `items` - List of scan logs. Each item contains:
    - `log_id` - Log entry ID.
    - `details` - Log details (JSON).
    - `created_at` - Timestamp.
- `total_count` - Total number of log entries.
