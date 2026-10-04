# prisma-airs_gateway_ai_providers schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Reads the upstream AI provider catalog without managing integrations. The catalog contains provider-family UUIDs for integration creation, not existing integration or workspace-provider UUIDs. No credentials are returned.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ids_by_slug` | `map(string)` | computed | — | Active provider-family UUIDs keyed by exact catalog slug. Duplicate or empty identities cause an error rather than an ambiguous lookup. |
| `items` | `list(object)` | computed | — | Provider definitions returned by the catalog, including inactive definitions. |

### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Provider-family UUID accepted as integration ai_provider_id. |
| `name` | `string` | computed | — | Catalog display name. |
| `slug` | `string` | computed | — | Exact catalog slug; OpenAI uses open-ai and Anthropic uses anthropic. |
| `status` | `string` | computed | — | Catalog lifecycle status. |
