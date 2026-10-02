# prisma-airs_runtime_deployment_profiles schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Fetches deployment profiles.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `items` | `list(object)` | computed | — | List of deployment profiles. |
| `limit` | `number` | optional | — | Maximum number of results to return. |
| `offset` | `number` | optional | — | Offset for pagination. |
| `total_count` | `number` | computed | — | Number of deployment profiles returned. |

### Attributes.items

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `auth_code` | `string` | computed | yes | Auth code for API key creation. |
| `details` | `string` | computed | yes | Profile details as JSON string. |
| `profile_id` | `string` | computed | yes | Sensitive legacy alias of auth_code; the API exposes no separate profile ID. |
| `profile_name` | `string` | computed | — | Deployment profile name. |
