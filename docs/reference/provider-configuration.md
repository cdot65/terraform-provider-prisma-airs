# Provider Configuration

Complete reference for all provider configuration attributes.

## Schema

```hcl
provider "prisma-airs" {
  # Scan API (API Key auth)
  api_key      = string  # optional, sensitive
  api_token    = string  # optional, sensitive
  profile_name = string  # optional
  endpoint     = string  # optional

  # OAuth2 (Management, Model Security, Red Team)
  client_id      = string  # optional
  client_secret  = string  # optional, sensitive
  tsg_id         = string  # optional
  mgmt_endpoint  = string  # optional
  token_endpoint = string  # optional

  # Model Security endpoints
  model_sec_data_endpoint = string  # optional
  model_sec_mgmt_endpoint = string  # optional

  # Red Team endpoints
  red_team_data_endpoint = string  # optional
  red_team_mgmt_endpoint = string  # optional
}
```

## Attribute Reference

### Scan API

| Attribute | Description | Env Var | Default |
|-----------|-------------|---------|---------|
| `api_key` | API key for scan operations | `PANW_AI_SEC_API_KEY` | — |
| `api_token` | Bearer token for scan operations | `PANW_AI_SEC_API_TOKEN` | — |
| `profile_name` | Default AI security profile name | `PANW_AI_SEC_PROFILE_NAME` | — |
| `endpoint` | Scan API endpoint | `PANW_AI_SEC_API_ENDPOINT` | `https://service.api.aisecurity.paloaltonetworks.com` |

### OAuth2

| Attribute | Description | Env Var | Default |
|-----------|-------------|---------|---------|
| `client_id` | OAuth2 client ID | `PANW_MGMT_CLIENT_ID` | — |
| `client_secret` | OAuth2 client secret | `PANW_MGMT_CLIENT_SECRET` | — |
| `tsg_id` | Tenant Service Group ID | `PANW_MGMT_TSG_ID` | — |
| `mgmt_endpoint` | Management API endpoint | `PANW_MGMT_ENDPOINT` | `https://api.sase.paloaltonetworks.com/aisec` |
| `token_endpoint` | OAuth2 token endpoint | `PANW_MGMT_TOKEN_ENDPOINT` | `https://auth.apps.paloaltonetworks.com/oauth2/access_token` |

### Model Security

| Attribute | Description | Env Var | Default |
|-----------|-------------|---------|---------|
| `model_sec_data_endpoint` | Data plane endpoint | `PANW_MODEL_SEC_DATA_ENDPOINT` | `https://api.sase.paloaltonetworks.com/aims/data` |
| `model_sec_mgmt_endpoint` | Management plane endpoint | `PANW_MODEL_SEC_MGMT_ENDPOINT` | `https://api.sase.paloaltonetworks.com/aims/mgmt` |

### Red Team

| Attribute | Description | Env Var | Default |
|-----------|-------------|---------|---------|
| `red_team_data_endpoint` | Data plane endpoint | `PANW_RED_TEAM_DATA_ENDPOINT` | `https://api.sase.paloaltonetworks.com/ai-red-teaming/data-plane` |
| `red_team_mgmt_endpoint` | Management plane endpoint | `PANW_RED_TEAM_MGMT_ENDPOINT` | `https://api.sase.paloaltonetworks.com/ai-red-teaming/mgmt-plane` |
