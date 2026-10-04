# Gateway AI provider catalog

Discover the provider-family UUIDs needed to create upstream integrations. Select a readable catalog slug instead of copying a UUID from the console or CLI.

This data source is new development functionality, pending the next provider release. Provider 0.10.0 does not expose it. For local use, build this provider checkout and use a [development override](../getting-started/installation.md#use-the-updated-provider-from-source).

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

These are **provider-family IDs**, not existing connection IDs. Use [Gateway integrations](gateway-integrations.md) to discover existing organization integrations and [Gateway providers](gateway-providers.md) for workspace providers. Those inventories return individual pages; the catalog endpoint has no pagination arguments in the published SDK.

See the complete [OpenAI and Claude Opus example](../examples/gateway.md) for connections, workspace bindings, and routing. Model access and API credentials remain upstream prerequisites; discovering the provider family does not discover or enable a model.
