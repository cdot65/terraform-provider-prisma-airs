# prisma-airs_red_team_quota

Reads quota information from Prisma AIRS Red Team API.

## Example Usage

```hcl
data "prisma-airs_red_team_quota" "current" {}

output "remaining_scans" {
  value = data.prisma-airs_red_team_quota.current.remaining
}
```

## Argument Reference

No arguments required.

## Attribute Reference

- `total` - Total quota allocated.
- `used` - Quota used.
- `remaining` - Quota remaining.
