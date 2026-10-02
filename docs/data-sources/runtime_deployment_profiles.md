---
page_title: "prisma-airs_runtime_deployment_profiles (Data Source)"
subcategory: "AI Runtime Security"
---

# prisma-airs_runtime_deployment_profiles Data Source

Reads deployment profiles from Prisma AIRS Management API.

## Example Usage

```hcl
data "prisma-airs_runtime_deployment_profiles" "all" {
  limit = 10
}

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

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_runtime_deployment_profiles/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `items` | `list(object)` | computed | — | List of deployment profiles. |
| `limit` | `number` | optional | — | Maximum number of results to return. |
| `offset` | `number` | optional | — | Offset for pagination. |
| `total_count` | `number` | computed | — | Number of deployment profiles returned. |

#### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `auth_code` | `string` | computed | yes | Auth code for API key creation. |
| `details` | `string` | computed | yes | Profile details as JSON string. |
| `profile_id` | `string` | computed | yes | Sensitive legacy alias of auth_code; the API exposes no separate profile ID. |
| `profile_name` | `string` | computed | — | Deployment profile name. |
