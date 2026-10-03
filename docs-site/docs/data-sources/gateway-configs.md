# AI Gateway configs

Reads `prisma-airs_gateway_configs` metadata without modifying remote objects.

## Example

```hcl
# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_gateway_configs" "example" {
  workspace_id = var.workspace_id
}
```

## Results

`items` contains typed identifiers, names, lifecycle status and available scope metadata. Revision IDs and boolean `is_default`/`enabled` metadata are included when returned; missing fields are null. Config documents, upstream credentials, API keys and deployment auth are never included. `total_count` uses the server total where available, otherwise the returned item count. Archived records may remain visible.

This route has no pagination arguments in the pinned SDK contract. It returns the service’s default page; do not treat it as an exhaustive inventory.

See the [exact schema reference](../reference/generated/prisma-airs_gateway_configs.md) and [Gateway workflow](../guides/gateway-workflow.md).
