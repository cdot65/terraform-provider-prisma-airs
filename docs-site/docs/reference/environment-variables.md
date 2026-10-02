# Environment variables

The provider reads these environment variables when their corresponding attributes are unset. It resolves one common OAuth identity and passes it to all service clients.

| Variable | Provider attribute | Purpose |
| --- | --- | --- |
| `PANW_MGMT_CLIENT_ID` | `client_id` | OAuth client ID |
| `PANW_MGMT_CLIENT_SECRET` | `client_secret` | OAuth client secret; sensitive |
| `PANW_MGMT_TSG_ID` | `tsg_id` | Tenant service group ID |
| `PANW_MGMT_ENDPOINT` | `mgmt_endpoint` | Management API override |
| `PANW_MGMT_TOKEN_ENDPOINT` | `token_endpoint` | Common OAuth token endpoint override |
| `PANW_MODEL_SEC_DATA_ENDPOINT` | `model_sec_data_endpoint` | Model Security data API override |
| `PANW_MODEL_SEC_MGMT_ENDPOINT` | `model_sec_mgmt_endpoint` | Model Security management API override |
| `PANW_RED_TEAM_DATA_ENDPOINT` | `red_team_data_endpoint` | Red Team data API override |
| `PANW_RED_TEAM_MGMT_ENDPOINT` | `red_team_mgmt_endpoint` | Red Team management API override |

An explicit provider attribute takes precedence over its corresponding variable. Service-specific SDK credential prefixes do not establish separate identities through this provider. Use [aliased configurations](../getting-started/configuration.md#separate-tenants) for distinct credentials.

## Local development

`TF_CLI_CONFIG_FILE` selects your Terraform CLI configuration, including a development override. See [Installation](../getting-started/installation.md).

Terraform does not load `.env` automatically. Repository examples may use `../../scripts/terraform-env.sh plan` to load a local file; CI should use its secret-store integration. Keep credentials and state out of source control.
