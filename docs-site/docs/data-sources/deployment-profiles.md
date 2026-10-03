# prisma-airs_runtime_deployment_profiles

Reads deployment profiles from Prisma AIRS Management API.

## Example Usage

```hcl
# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_runtime_deployment_profiles" "all" {
  limit = 10
}

# Outputs: Expose results for the next configuration or application step.
output "profiles" {
  value = [for p in data.prisma-airs_runtime_deployment_profiles.all.items : p.profile_name]
}
```

## Argument Reference

- `limit` - (Optional) Maximum number of profiles to return.
- `offset` - (Optional) Offset for pagination.

## Attribute Reference

- `items` - List of deployment profiles. Each item contains:
    - `profile_id` - Sensitive legacy alias of `auth_code`; no separate ID is supplied by the API.
    - `profile_name` - Deployment profile name.
    - `auth_code` - Sensitive auth code for API key creation.
    - `details` - Sensitive full profile details as a JSON string.
- `total_count` - Number of profiles returned.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_runtime_deployment_profiles.md) for all nested fields, types, and sensitivity flags.
