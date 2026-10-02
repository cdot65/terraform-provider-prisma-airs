# prisma-airs_runtime_dlp_profiles schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Fetches DLP profiles.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `items` | `list(object)` | computed | — | List of DLP profiles. |
| `limit` | `number` | optional | — | Maximum number of results to return. |
| `offset` | `number` | optional | — | Offset for pagination. |
| `total_count` | `number` | computed | — | Number of DLP profiles returned. |

### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `details` | `string` | computed | — | Profile details as JSON string. |
| `profile_id` | `string` | computed | — | DLP profile ID. |
| `profile_name` | `string` | computed | — | DLP profile name. |
