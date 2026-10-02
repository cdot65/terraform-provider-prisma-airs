# Provider schema

Exact provider attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Terraform provider for Palo Alto Networks Prisma AI Runtime Security (AIRS).

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `client_id` | `string` | optional | — | OAuth2 client ID for management APIs. Can also be set via PANW_MGMT_CLIENT_ID. |
| `client_secret` | `string` | optional | yes | OAuth2 client secret for management APIs. Can also be set via PANW_MGMT_CLIENT_SECRET. |
| `mgmt_endpoint` | `string` | optional | — | Management API endpoint override. Can also be set via PANW_MGMT_ENDPOINT. |
| `model_sec_data_endpoint` | `string` | optional | — | Model Security data plane endpoint. Can also be set via PANW_MODEL_SEC_DATA_ENDPOINT. |
| `model_sec_mgmt_endpoint` | `string` | optional | — | Model Security management plane endpoint. Can also be set via PANW_MODEL_SEC_MGMT_ENDPOINT. |
| `red_team_data_endpoint` | `string` | optional | — | Red Team data plane endpoint. Can also be set via PANW_RED_TEAM_DATA_ENDPOINT. |
| `red_team_mgmt_endpoint` | `string` | optional | — | Red Team management plane endpoint. Can also be set via PANW_RED_TEAM_MGMT_ENDPOINT. |
| `token_endpoint` | `string` | optional | — | OAuth2 token endpoint override. Can also be set via PANW_MGMT_TOKEN_ENDPOINT. |
| `tsg_id` | `string` | optional | — | Tenant Service Group ID. Can also be set via PANW_MGMT_TSG_ID. |
