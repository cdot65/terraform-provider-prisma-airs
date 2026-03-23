# Configuration

The provider uses OAuth2 client credentials for authentication across all service domains.

## Provider Block

```hcl
provider "prisma-airs" {
  client_id     = var.panw_client_id
  client_secret = var.panw_client_secret
  tsg_id        = var.panw_tsg_id
}
```

All attributes support environment variable fallback — see [Environment Variables](../reference/environment-variables.md).

## Authentication

### OAuth2 Client Credentials

Used for all operations (Management, Model Security, Red Team).

```hcl
provider "prisma-airs" {
  client_id     = var.panw_client_id
  client_secret = var.panw_client_secret
  tsg_id        = var.panw_tsg_id
}
```

The provider handles the full OAuth2 token lifecycle automatically — token acquisition, caching, proactive refresh, and 401/403 auto-retry.

## Endpoint Overrides

For non-default regions or custom deployments:

```hcl
provider "prisma-airs" {
  client_id     = var.panw_client_id
  client_secret = var.panw_client_secret
  tsg_id        = var.panw_tsg_id

  # Management API
  mgmt_endpoint  = "https://api.sase.paloaltonetworks.com/aisec"
  token_endpoint  = "https://auth.apps.paloaltonetworks.com/oauth2/access_token"

  # Model Security (separate data/mgmt planes)
  model_sec_data_endpoint = "https://api.sase.paloaltonetworks.com/aims/data"
  model_sec_mgmt_endpoint = "https://api.sase.paloaltonetworks.com/aims/mgmt"

  # Red Team (separate data/mgmt planes)
  red_team_data_endpoint = "https://api.sase.paloaltonetworks.com/ai-red-teaming/data-plane"
  red_team_mgmt_endpoint = "https://api.sase.paloaltonetworks.com/ai-red-teaming/mgmt-plane"
}
```

## Using Variables

Best practice is to use variables for sensitive values:

```hcl
variable "panw_client_id" {
  type      = string
  sensitive = true
}

variable "panw_client_secret" {
  type      = string
  sensitive = true
}

variable "panw_tsg_id" {
  type = string
}
```

Or rely entirely on environment variables for CI/CD:

```bash
export PANW_MGMT_CLIENT_ID=your-client-id
export PANW_MGMT_CLIENT_SECRET=your-client-secret
export PANW_MGMT_TSG_ID=1234567890
```

## Using .env Files

For local development, store credentials in a `.env` file instead of exporting them manually:

```bash
# .env
PANW_MGMT_CLIENT_ID=your-client-id
PANW_MGMT_CLIENT_SECRET=your-client-secret
PANW_MGMT_TSG_ID=1234567890
```

Then either source it before running Terraform:

```bash
source .env
terraform plan
```

Or use the provided helper script which loads `.env` automatically:

```bash
../../scripts/terraform-env.sh plan
../../scripts/terraform-env.sh apply
```

**Important:** Never commit `.env` files — add them to `.gitignore`.
