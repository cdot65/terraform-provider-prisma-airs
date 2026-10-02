# prisma-airs_supply_chain_security_rules schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Fetches the catalog of Model Security rules.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `rules` | `list(object)` | computed | — | List of security rules. |

### Attributes.rules

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | computed | — | Rule description. |
| `name` | `string` | computed | — | Rule name. |
| `rule_type` | `string` | computed | — | Rule type (METADATA, ARTIFACT). |
| `source_type` | `string` | computed | — | Source type. |
| `uuid` | `string` | computed | — | Rule UUID. |
