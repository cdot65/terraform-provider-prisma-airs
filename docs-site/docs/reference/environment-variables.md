# Environment variables

The provider reads these environment variables when their corresponding attributes are unset. It resolves one common OAuth identity and passes it to all service clients.

| Variable | Provider attribute | Purpose |
| --- | --- | --- |
| `PANW_MGMT_CLIENT_ID` | `client_id` | OAuth client ID |
| `PANW_MGMT_CLIENT_SECRET` | `client_secret` | OAuth client secret; sensitive |
| `PANW_MGMT_TSG_ID` | `tsg_id` | Tenant service group ID |
| `PANW_MGMT_ENDPOINT` | `runtime.mgmt_endpoint` | Management API override |
| `PANW_MGMT_TOKEN_ENDPOINT` | `token_endpoint` | Common OAuth token endpoint override |
| `PANW_MODEL_SEC_DATA_ENDPOINT` | `supply_chain.data_endpoint` | Supply Chain Security model data API override |
| `PANW_MODEL_SEC_MGMT_ENDPOINT` | `supply_chain.mgmt_endpoint` | Supply Chain Security model management API override |
| `PANW_AI_GW_DATA_ENDPOINT` | `gateway.data_endpoint` | Gateway data-plane management override |
| `PANW_AI_GW_ADMIN_ENDPOINT` | `gateway.admin_endpoint` | Gateway admin-plane management override |
| `PANW_IAM_ENDPOINT` | `gateway.iam_endpoint` | SCM IAM override for workspace scope orchestration |
| `PANW_RED_TEAM_DATA_ENDPOINT` | `red_team.data_endpoint` | Red Team data API override |
| `PANW_RED_TEAM_MGMT_ENDPOINT` | `red_team.mgmt_endpoint` | Red Team management API override |

An explicit provider attribute takes precedence over its corresponding variable. Service-specific SDK credential prefixes do not establish separate identities through this provider. Use [aliased configurations](../getting-started/configuration.md#separate-tenants) for distinct credentials.

## Local development

`TF_CLI_CONFIG_FILE` selects your Terraform CLI configuration, including a development override. See [Installation](../getting-started/installation.md).

Terraform does not load `.env` automatically. Repository examples may use `../../scripts/terraform-env.sh plan` to load a local file; CI should use its secret-store integration. Keep credentials and state out of source control.

## Skill Scanning

`PANW_SKILL_SCANNING_DATA_ENDPOINT` and `PANW_SKILL_SCANNING_MGMT_ENDPOINT` configure the two bases under `supply_chain`. Both are required for Skill Scanning resources and data sources. Explicit `skill_scanning_data_endpoint` / `skill_scanning_mgmt_endpoint` settings take precedence. Established `PANW_AGENT_GUARD_*_ENDPOINT` SDK variables remain fallback aliases; shared `PANW_MGMT_*` credentials are used.
