# prisma-airs_runtime_dlp_profiles

Reads DLP data profiles from Prisma AIRS Management API.

## Example Usage

```hcl
# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_runtime_dlp_profiles" "all" {}

# Outputs: Expose results for the next configuration or application step.
output "profile_count" {
  value = data.prisma-airs_runtime_dlp_profiles.all.total_count
}
```

## Argument Reference

- `limit` - (Optional) Maximum number of profiles to return.
- `offset` - (Optional) Offset for pagination.

## Attribute Reference

- `items` - List of DLP profiles. Each item contains:
    - `profile_id` - Profile ID.
    - `profile_name` - Profile name.
    - `details` - Profile details (JSON).
- `total_count` - Total number of profiles.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_runtime_dlp_profiles.md) for all nested fields, types, and sensitivity flags.
