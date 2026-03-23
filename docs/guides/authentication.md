# Authentication

The Prisma AIRS Terraform provider uses OAuth2 client credentials for authentication across all service domains.

## OAuth2 Authentication

Used for all operations (Management, Model Security, Red Team).

### Provider Configuration

```hcl
provider "prisma-airs" {
  client_id     = var.panw_client_id
  client_secret = var.panw_client_secret
  tsg_id        = var.panw_tsg_id
}
```

### Environment Variables

```bash
export PANW_MGMT_CLIENT_ID=your-client-id
export PANW_MGMT_CLIENT_SECRET=your-client-secret
export PANW_MGMT_TSG_ID=1234567890
```

### Token Lifecycle

The provider handles the full OAuth2 token lifecycle automatically:

1. **Token acquisition** — fetches a token on first API call
2. **Token caching** — reuses the token for subsequent calls
3. **Proactive refresh** — refreshes 30 seconds before expiry
4. **Auto-retry** — retries on 401/403 with a fresh token

### Service-Specific Credentials

Model Security and Red Team can use their own credentials, falling back to the shared `PANW_MGMT_*` variables:

```bash
# Model Security specific (optional)
export PANW_MODEL_SEC_CLIENT_ID=...
export PANW_MODEL_SEC_CLIENT_SECRET=...
export PANW_MODEL_SEC_TSG_ID=...

# Red Team specific (optional)
export PANW_RED_TEAM_CLIENT_ID=...
export PANW_RED_TEAM_CLIENT_SECRET=...
export PANW_RED_TEAM_TSG_ID=...
```

## Using .env Files

For local development, store credentials in a `.env` file instead of exporting them in your shell:

```bash
# .env
PANW_MGMT_CLIENT_ID=your-client-id
PANW_MGMT_CLIENT_SECRET=your-client-secret
PANW_MGMT_TSG_ID=1234567890
```

Load it manually or use the repo helper:

```bash
source .env && terraform plan
# or
../../scripts/terraform-env.sh plan
```

Never commit `.env` files to version control.

## Precedence

Provider config attributes always take precedence over environment variables:

1. Explicit provider attribute value
2. Service-specific environment variable (e.g., `PANW_MODEL_SEC_CLIENT_ID`)
3. Fallback environment variable (e.g., `PANW_MGMT_CLIENT_ID`)
