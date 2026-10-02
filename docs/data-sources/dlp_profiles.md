---
page_title: "prisma-airs_dlp_profiles (Data Source)"
---

# prisma-airs_dlp_profiles Data Source

Reads DLP data profiles from Prisma AIRS Management API.

## Example Usage

```hcl
data "prisma-airs_dlp_profiles" "all" {}

output "profile_count" {
  value = data.prisma-airs_dlp_profiles.all.total_count
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

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_dlp_profiles/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `items` | `list(object)` | computed | — | List of DLP profiles. |
| `limit` | `number` | optional | — | Maximum number of results to return. |
| `offset` | `number` | optional | — | Offset for pagination. |
| `total_count` | `number` | computed | — | Number of DLP profiles returned. |

#### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `details` | `string` | computed | — | Profile details as JSON string. |
| `profile_id` | `string` | computed | — | DLP profile ID. |
| `profile_name` | `string` | computed | — | DLP profile name. |
