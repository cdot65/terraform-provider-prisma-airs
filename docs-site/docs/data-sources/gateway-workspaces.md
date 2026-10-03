# AI Gateway workspace discovery

`prisma-airs_gateway_workspaces` lists safe admin-plane workspace metadata.

```hcl
# Discovery: List the active workspace inventory.
data "prisma-airs_gateway_workspaces" "active" {
  status = "active"
}
```

Use `status = "archived"` for archived rows. Results expose `items`, `total_count`, `has_more` and `complete`. An omitted `has_more` stays null. `complete` is true only when an explicit false flag and the reported total match the returned records. Incomplete results produce a warning and cannot prove absence or unique scope ownership.

The captured route has no pagination parameters. This data source does not invent page/offset controls or exhaust pages it cannot request. Credentials, defaults, users and security settings are excluded from every item.

See the [exact list schema](../reference/generated/prisma-airs_gateway_workspaces.md) and [workspace lookup](gateway-workspace.md).

List rows retain their literal display labels, including a label that begins with its icon text. Only the singular detail endpoint decorates its returned label.
