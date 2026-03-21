# prisma-airs_deployment_profiles

Reads deployment profiles from Prisma AIRS Management API.

## Example Usage

```hcl
data "prisma-airs_deployment_profiles" "all" {}

output "profiles" {
  value = data.prisma-airs_deployment_profiles.all.items
}
```

## Argument Reference

- `limit` - (Optional) Maximum number of profiles to return.
- `offset` - (Optional) Offset for pagination.

## Attribute Reference

- `items` - List of deployment profiles. Each item contains:
    - `profile_id` - Profile ID.
    - `profile_name` - Profile name.
    - `details` - Profile details (JSON).
- `total_count` - Total number of profiles.
