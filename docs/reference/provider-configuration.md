# Provider Configuration

Complete reference for all provider configuration attributes.

## Schema

```hcl
provider "prisma-airs" {
  # OAuth2 (all API domains)
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

## Minimal Example

```hcl
# All credentials via environment variables
provider "prisma-airs" {}
```

## Full Example

```hcl
provider "prisma-airs" {
  client_id     = var.panw_client_id
  client_secret = var.panw_client_secret
  tsg_id        = var.panw_tsg_id
}
```
