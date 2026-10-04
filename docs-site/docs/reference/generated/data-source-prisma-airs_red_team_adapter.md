# prisma-airs_red_team_adapter schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Reads an existing Red Team adapter without managing or executing it. Secret values remain null.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `description` | `string` | computed | — | Adapter description. |
| `id` | `string` | required | — | Adapter UUID, typically selected from red_team_adapters.ids_by_name. |
| `name` | `string` | computed | — | Adapter name. |
| `network_broker_channel_uuid` | `string` | computed | — | Existing channel reference, if present. |
| `script` | `string` | computed | yes | Decoded Python script; never executed by this data source. |
| `status` | `string` | computed | — | DRAFT or ACTIVE. |
| `variables` | `map(object)` | computed | yes | Variable key/type inventory; redacted secrets have null values. |

### Attributes.variables

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `type` | `string` | computed | — | VAR or SECRET. |
| `value` | `string` | computed | yes | Visible value or null if redacted. |
