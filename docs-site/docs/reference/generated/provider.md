# Provider schema

Exact provider attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Terraform provider for Prisma AIRS: AI Runtime Security, AI Red Teaming, AI Gateway, and AI Supply Chain Security.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `client_id` | `string` | optional | — | Shared OAuth2 client ID. Can also be set via PANW_MGMT_CLIENT_ID. |
| `client_secret` | `string` | optional | yes | Shared OAuth2 client secret. Can also be set via PANW_MGMT_CLIENT_SECRET. |
| `token_endpoint` | `string` | optional | — | Shared OAuth2 token endpoint override. Can also be set via PANW_MGMT_TOKEN_ENDPOINT. |
| `tsg_id` | `string` | optional | — | Tenant Service Group ID. Can also be set via PANW_MGMT_TSG_ID. |

## gateway

Nesting: `single`.

### gateway

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `admin_endpoint` | `string` | optional | — | Gateway admin-plane management endpoint override. Can also be set via PANW_AI_GW_ADMIN_ENDPOINT. |
| `data_endpoint` | `string` | optional | — | Gateway data-plane management endpoint override. Can also be set via PANW_AI_GW_DATA_ENDPOINT. |

## red_team

Nesting: `single`.

### red_team

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `data_endpoint` | `string` | optional | — | Red Teaming data endpoint override. Can also be set via PANW_RED_TEAM_DATA_ENDPOINT. |
| `mgmt_endpoint` | `string` | optional | — | Red Teaming management endpoint override. Can also be set via PANW_RED_TEAM_MGMT_ENDPOINT. |

## runtime

Nesting: `single`.

### runtime

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `mgmt_endpoint` | `string` | optional | — | Runtime Security management endpoint override. Can also be set via PANW_MGMT_ENDPOINT. |

## supply_chain

Nesting: `single`.

### supply_chain

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `data_endpoint` | `string` | optional | — | Supply Chain Security model-management data endpoint override. Can also be set via PANW_MODEL_SEC_DATA_ENDPOINT. |
| `mgmt_endpoint` | `string` | optional | — | Supply Chain Security model-management endpoint override. Can also be set via PANW_MODEL_SEC_MGMT_ENDPOINT. |
| `skill_scanning_data_endpoint` | `string` | optional | — | Skill Scanning data base URL, including its product prefix. Both Skill Scanning endpoints are required when using its resources. Can also be set via PANW_SKILL_SCANNING_DATA_ENDPOINT. |
| `skill_scanning_mgmt_endpoint` | `string` | optional | — | Skill Scanning management base URL, including its product prefix. Can also be set via PANW_SKILL_SCANNING_MGMT_ENDPOINT. |
