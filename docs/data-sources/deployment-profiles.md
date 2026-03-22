# prisma-airs_deployment_profiles

Reads deployment profiles from Prisma AIRS Management API.

## Example Usage

```hcl
data "prisma-airs_deployment_profiles" "all" {
  limit = 10
}

output "profiles" {
  value = [for p in data.prisma-airs_deployment_profiles.all.items : p.profile_name]
}
```

## Argument Reference

- `limit` - (Optional) Maximum number of profiles to return.
- `offset` - (Optional) Offset for pagination.

## Attribute Reference

- `items` - List of deployment profiles. Each item contains:
    - `profile_id` - Deployment profile ID (same value as `auth_code`).
    - `profile_name` - Deployment profile name.
    - `auth_code` - Auth code for API key creation.
    - `details` - Full profile details as a JSON string.
- `total_count` - Number of profiles returned.
