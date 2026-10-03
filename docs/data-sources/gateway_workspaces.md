---
page_title: "prisma-airs_gateway_workspaces (Data Source)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_workspaces Data Source

`prisma-airs_gateway_workspaces` lists safe admin-plane workspace metadata.

```hcl
# Discovery: List the active workspace inventory.
data "prisma-airs_gateway_workspaces" "active" {
  status = "active"
}
```

Use `status = "archived"` for archived rows. Results expose `items`, `total_count`, `has_more` and `complete`. An omitted `has_more` stays null. `complete` is true only when an explicit false flag and the reported total match the returned records. Incomplete results produce a warning and cannot prove absence or unique scope ownership.

The captured route has no pagination parameters. This data source does not invent page/offset controls or exhaust pages it cannot request. Credentials, defaults, users and security settings are excluded from every item.

See the [exact list schema](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_workspaces/) and [workspace lookup](https://cdot65.github.io/terraform-provider-prisma-airs/data-sources/gateway-workspace/).

List rows retain their literal display labels, including a label that begins with its icon text. Only the singular detail endpoint decorates its returned label.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `complete` | `bool` | computed | — | True only when the response explicitly establishes a complete inventory. |
| `has_more` | `bool` | computed | — | Whether more records are reported; null if the response omits this flag. |
| `items` | `list(object)` | computed | — | Safe metadata for the returned workspace page. |
| `status` | `string` | optional | — | Lifecycle filter; defaults to active. |
| `total_count` | `number` | computed | — | Reported inventory total. |

#### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Remote workspace created_at when available. |
| `description` | `string` | computed | — | Remote workspace description when available. |
| `id` | `string` | computed | — | Remote workspace id when available. |
| `is_default` | `bool` | computed | — | Whether this is the tenant default workspace. |
| `last_updated_at` | `string` | computed | — | Remote workspace last_updated_at when available. |
| `name` | `string` | computed | — | Remote workspace name when available. |
| `scope_name` | `string` | computed | — | Remote workspace scope_name when available. |
| `slug` | `string` | computed | — | Remote workspace slug when available. |
| `status` | `string` | computed | — | Remote workspace status when available. |
