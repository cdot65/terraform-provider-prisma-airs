---
page_title: "prisma-airs_gateway_ai_providers (Data Source)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_ai_providers Data Source

Discover the provider-family UUIDs needed to create upstream integrations. Select a readable catalog slug instead of copying a UUID from the console or CLI.

This data source is new development functionality, pending the next provider release. Provider 0.10.0 does not expose it. For local use, build this provider checkout and use a [development override](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/installation/#use-the-updated-provider-from-source).

```hcl
# Discovery: Read the upstream catalog without owning its entries.
data "prisma-airs_gateway_ai_providers" "catalog" {}

# Connection: Resolve OpenAI's family UUID by its actual catalog slug.
resource "prisma-airs_gateway_integration" "openai" {
  name           = "Example OpenAI"
  ai_provider_id = data.prisma-airs_gateway_ai_providers.catalog.ids_by_slug["open-ai"]
  key            = var.openai_api_key
}
```

`items` contains typed `id`, `slug`, `name`, and `status` metadata for every entry returned by the catalog endpoint. `ids_by_slug` is a computed `map(string)` containing active entries only. The exact OpenAI slug is **`open-ai`**; Anthropic is **`anthropic`**. Neither `chatgpt` nor a Claude model name is a provider-family slug.

Read a UUID with `data.prisma-airs_gateway_ai_providers.catalog.ids_by_slug["anthropic"]`, pass the map to a module, or output it for inspection. Terraform refreshes this read-only data; it does not create, update, or delete catalog entries. Credentials and unrecognized API fields are excluded from state. Duplicate or empty IDs/slugs produce an error; a missing or inactive slug makes a direct map lookup fail rather than select another service.

These are **provider-family IDs**, not existing connection IDs. Use [Gateway integrations](https://cdot65.github.io/terraform-provider-prisma-airs/data-sources/gateway-integrations/) to discover existing organization integrations and [Gateway providers](https://cdot65.github.io/terraform-provider-prisma-airs/data-sources/gateway-providers/) for workspace providers. Those inventories return individual pages; the catalog endpoint has no pagination arguments in the published SDK.

See the complete [OpenAI and Claude Opus example](https://cdot65.github.io/terraform-provider-prisma-airs/examples/gateway/) for connections, workspace bindings, and routing. Model access and API credentials remain upstream prerequisites; discovering the provider family does not discover or enable a model.

## Run discovery before configuring models

You can read the catalog without an upstream API key. Save this complete configuration in a separate directory and use the development override described above:

```hcl
# Setup: Use the catalog-capable development provider until its release.
terraform {
  required_version = ">= 1.11.0, < 2.0.0"

  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "> 0.10.0, < 1.0.0"
    }
  }
}

# Authentication: Discovery needs management credentials, not upstream API keys.
provider "prisma-airs" {}

# Discovery: Read catalog entries without owning upstream connections.
data "prisma-airs_gateway_ai_providers" "catalog" {}

# Lookup: Select the UUIDs by the exact catalog slugs used for integration creation.
output "provider_family_ids" {
  value = {
    open-ai   = data.prisma-airs_gateway_ai_providers.catalog.ids_by_slug["open-ai"]
    anthropic = data.prisma-airs_gateway_ai_providers.catalog.ids_by_slug["anthropic"]
  }
}

output "catalog_counts" {
  value = {
    returned = length(data.prisma-airs_gateway_ai_providers.catalog.items)
    active   = length(data.prisma-airs_gateway_ai_providers.catalog.ids_by_slug)
  }
}
```

Run `terraform apply`, then `terraform output -no-color`. An actual run on 2026-10-04 returned the following; provider UUID values are sanitized, while Terraform's output structure and counts are retained:

```text
catalog_counts = {
  "active" = 79
  "returned" = 81
}
provider_family_ids = {
  "anthropic" = "<anthropic-provider-family-uuid>"
  "open-ai" = "<openai-provider-family-uuid>"
}
```

The apply reported `Resources: 0 added, 0 changed, 0 destroyed.` Its subsequent unchanged plan exited 0, and cleanup left no state entries. These results come from a live catalog read, not mock data. Your catalog counts and UUIDs can differ.

See the [recorded CLI transcript](https://github.com/cdot65/prisma-airs-terraform-examples/blob/feat/gateway-catalog-examples/docs/live-runs/gateway-provider-catalog.txt) and [source/build receipt](https://github.com/cdot65/prisma-airs-terraform-examples/blob/feat/gateway-catalog-examples/docs/live-runs/gateway-provider-catalog-receipt.json). Run `terraform destroy` to remove the read-only state outputs when finished; no catalog entry is deleted.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ids_by_slug` | `map(string)` | computed | — | Active provider-family UUIDs keyed by exact catalog slug. Duplicate or empty identities cause an error rather than an ambiguous lookup. |
| `items` | `list(object)` | computed | — | Provider definitions returned by the catalog, including inactive definitions. |

#### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Provider-family UUID accepted as integration ai_provider_id. |
| `name` | `string` | computed | — | Catalog display name. |
| `slug` | `string` | computed | — | Exact catalog slug; OpenAI uses open-ai and Anthropic uses anthropic. |
| `status` | `string` | computed | — | Catalog lifecycle status. |
