# prisma-airs_red_team_adapters

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

Use the [single adapter data source](red-team-adapter.md) to read script and variable metadata. Discovery never creates or deletes adapters; import an adapter resource to adopt ownership.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_red_team_adapters.md).
