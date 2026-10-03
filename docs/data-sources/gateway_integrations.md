---
page_title: "prisma-airs_gateway_integrations (Data Source)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_integrations Data Source

Reads `prisma-airs_gateway_integrations` metadata without modifying remote objects.

## Example

```hcl
data "prisma-airs_gateway_integrations" "example" {
  page_size = 100
  current_page = 1
}
```

## Results

`items` contains typed identifiers, names, lifecycle status and available scope metadata. Missing fields are null. Config documents, upstream credentials, API keys and deployment auth are never included. `total_count` uses the server total where available, otherwise the returned item count. Archived records may remain visible.

This returns one page. `current_page` is one-based and `page_size` defaults to 100; request further pages explicitly. A page is not an exhaustive inventory.

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_integrations/) and [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/).

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `current_page` | `number` | optional | — | One-based page index (default 1). |
| `items` | `list(object)` | computed | — | Safe resource metadata for the returned page. Credentials and config documents are never included. |
| `page_size` | `number` | optional | — | Results per page (default 100). |
| `total_count` | `number` | computed | — | Server-reported total where available; otherwise the number of returned items. |

#### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ai_provider_id` | `string` | computed | — | Remote ai_provider_id when available; otherwise null. |
| `created_at` | `string` | computed | — | Remote created_at when available; otherwise null. |
| `id` | `string` | computed | — | Remote id when available; otherwise null. |
| `integration_id` | `string` | computed | — | Remote integration_id when available; otherwise null. |
| `last_updated_at` | `string` | computed | — | Remote last_updated_at when available; otherwise null. |
| `mcp_integration_id` | `string` | computed | — | Remote mcp_integration_id when available; otherwise null. |
| `name` | `string` | computed | — | Remote name when available; otherwise null. |
| `organisation_id` | `string` | computed | — | Remote organisation_id when available; otherwise null. |
| `slug` | `string` | computed | — | Remote slug when available; otherwise null. |
| `status` | `string` | computed | — | Remote status when available; otherwise null. |
| `type` | `string` | computed | — | Remote type when available; otherwise null. |
| `user_id` | `string` | computed | — | Remote user_id when available; otherwise null. |
| `workspace_id` | `string` | computed | — | Remote workspace_id when available; otherwise null. |
