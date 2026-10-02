# prisma-airs_runtime_api_key schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages an AI Runtime Security API key.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_key` | `string` | computed | yes | The API key value. Only available at creation time. |
| `api_key_id` | `string` | computed | — | The unique identifier of the API key. |
| `api_key_name` | `string` | required | — | Name of the API key. |
| `auth_code` | `string` | required | yes | Deployment profile auth code for API key creation. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `created_by` | `string` | optional, computed | — | Identity of the user creating the key. |
| `cust_ai_agent_framework` | `string` | optional, computed | — | Customer AI agent framework. |
| `cust_app` | `string` | optional, computed | — | Customer application name to associate with the key. |
| `cust_cloud_provider` | `string` | optional, computed | — | Customer cloud provider used when creating the associated application. |
| `cust_env` | `string` | optional, computed | — | Customer environment used when creating the associated application. |
| `expires_at` | `string` | computed | — | Expiration timestamp. |
| `id` | `string` | computed | — | Terraform resource ID (same as api_key_id). |
| `revoked` | `bool` | computed | — | Whether the API key is revoked. |
| `rotation_time_interval` | `number` | required | — | Rotation interval value (e.g. 90 for 90 days). |
| `rotation_time_unit` | `string` | required | — | Rotation time unit (days, months). |
| `status` | `string` | computed | — | API key status. |
