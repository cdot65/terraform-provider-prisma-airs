---
page_title: "prisma-airs_red_team_adapters (Data Source)"
subcategory: "AI Red Teaming"
---

# prisma-airs_red_team_adapters Data Source

Discover existing adapter UUIDs without taking ownership or executing scripts.

```hcl
# Discovery: Resolve an adapter by its exact tenant-visible name.
data "prisma-airs_red_team_adapters" "available" {}

resource "prisma-airs_red_team_target" "application" {
  name                        = "terraform-adapter-application"
  target_type                 = "APPLICATION"
  api_endpoint_type           = "NETWORK_BROKER"
  network_broker_channel_uuid = var.network_broker_channel_uuid

  adapter {
    uuid = data.prisma-airs_red_team_adapters.available.ids_by_name[var.adapter_name]
  }
}
```

`items` contains identity, name, and status for each adapter; `ids_by_name` provides exact name lookup. Pagination is bounded to 10,000 adapters and a three-minute timeout. Incomplete pagination and duplicate names/identities fail rather than publish an ambiguous result. A missing name fails the lookup: check spelling, tenant selection, and permissions.

Use the [single adapter data source](https://cdot65.github.io/terraform-provider-prisma-airs/data-sources/red-team-adapter/) to read script and variable metadata. Discovery never creates or deletes adapters; import an adapter resource to adopt ownership.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_red_team_adapters/).

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ids_by_name` | `map(string)` | computed | — | Adapter UUIDs keyed by exact name. Duplicate names fail rather than silently selecting an adapter. |
| `items` | `list(object)` | computed | — | Adapter identity and lifecycle metadata. |

#### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Adapter UUID. |
| `name` | `string` | computed | — | Adapter name. |
| `status` | `string` | computed | — | DRAFT or ACTIVE. |
