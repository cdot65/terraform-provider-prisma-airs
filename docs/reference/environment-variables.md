# Environment Variables

All provider configuration attributes can be set via environment variables.

## Management API (OAuth2)

| Variable | Required | Description |
|----------|----------|-------------|
| `PANW_MGMT_CLIENT_ID` | For management operations | OAuth2 client ID |
| `PANW_MGMT_CLIENT_SECRET` | For management operations | OAuth2 client secret |
| `PANW_MGMT_TSG_ID` | For management operations | Tenant Service Group ID |
| `PANW_MGMT_ENDPOINT` | Optional | Management API endpoint override |
| `PANW_MGMT_TOKEN_ENDPOINT` | Optional | OAuth2 token endpoint override |

## Model Security API

Falls back to `PANW_MGMT_*` variables if not set.

| Variable | Required | Description |
|----------|----------|-------------|
| `PANW_MODEL_SEC_CLIENT_ID` | Optional | Client ID (falls back to `PANW_MGMT_CLIENT_ID`) |
| `PANW_MODEL_SEC_CLIENT_SECRET` | Optional | Client secret (falls back to `PANW_MGMT_CLIENT_SECRET`) |
| `PANW_MODEL_SEC_TSG_ID` | Optional | TSG ID (falls back to `PANW_MGMT_TSG_ID`) |
| `PANW_MODEL_SEC_DATA_ENDPOINT` | Optional | Data plane endpoint |
| `PANW_MODEL_SEC_MGMT_ENDPOINT` | Optional | Management plane endpoint |

## Red Team API

Falls back to `PANW_MGMT_*` variables if not set.

| Variable | Required | Description |
|----------|----------|-------------|
| `PANW_RED_TEAM_CLIENT_ID` | Optional | Client ID (falls back to `PANW_MGMT_CLIENT_ID`) |
| `PANW_RED_TEAM_CLIENT_SECRET` | Optional | Client secret (falls back to `PANW_MGMT_CLIENT_SECRET`) |
| `PANW_RED_TEAM_TSG_ID` | Optional | TSG ID (falls back to `PANW_MGMT_TSG_ID`) |
| `PANW_RED_TEAM_DATA_ENDPOINT` | Optional | Data plane endpoint |
| `PANW_RED_TEAM_MGMT_ENDPOINT` | Optional | Management plane endpoint |

## Using .env Files

For local development, you can store credentials in a `.env` file:

```bash
# .env
PANW_MGMT_CLIENT_ID=your-client-id
PANW_MGMT_CLIENT_SECRET=your-client-secret
PANW_MGMT_TSG_ID=1234567890
```

The repo includes a helper at `scripts/terraform-env.sh` that loads `.env` before running Terraform:

```bash
../../scripts/terraform-env.sh plan
```

Never commit `.env` files to version control.

## Precedence

1. Provider block attribute (explicit value)
2. Service-specific environment variable
3. Fallback `PANW_MGMT_*` environment variable
