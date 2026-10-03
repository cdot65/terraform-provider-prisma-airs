# AI Gateway mcp integrations

Reads `prisma-airs_gateway_mcp_integrations` metadata without modifying remote objects.

## Example

```hcl
data "prisma-airs_gateway_mcp_integrations" "example" {
  page_size = 100
  current_page = 1
}
```

## Results

`items` contains typed identifiers, names, lifecycle status and available scope metadata. Revision IDs and boolean `is_default`/`enabled` metadata are included when returned; missing fields are null. Config documents, upstream credentials, API keys and deployment auth are never included. `total_count` uses the server total where available, otherwise the returned item count. Archived records may remain visible.

This returns one page. `current_page` is one-based and `page_size` defaults to 100; request further pages explicitly. A page is not an exhaustive inventory.

See the [exact schema reference](../reference/generated/prisma-airs_gateway_mcp_integrations.md) and [Gateway workflow](../guides/gateway-workflow.md).
