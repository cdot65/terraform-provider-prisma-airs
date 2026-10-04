# prisma-airs_red_team_adapters schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Discovers all Red Team adapters with bounded pagination. Reading never executes scripts or takes ownership.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ids_by_name` | `map(string)` | computed | — | Adapter UUIDs keyed by exact name. Duplicate names fail rather than silently selecting an adapter. |
| `items` | `list(object)` | computed | — | Adapter identity and lifecycle metadata. |

### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Adapter UUID. |
| `name` | `string` | computed | — | Adapter name. |
| `status` | `string` | computed | — | DRAFT or ACTIVE. |
