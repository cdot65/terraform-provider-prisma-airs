# Authentication

The Prisma AIRS Terraform provider supports two authentication methods, depending on which service domain you're using.

## Authentication Methods

| Method | Service Domains | Mechanism |
|--------|----------------|-----------|
| **API Key** | AI Runtime Security (Scan) | HMAC-SHA256 request signing |
| **OAuth2** | Management, Model Security, Red Team | Client credentials grant |

## API Key Authentication

Used exclusively for content scanning via the Scan API.

### Provider Configuration

```hcl
provider "prisma-airs" {
  api_key      = var.panw_api_key
  profile_name = "my-security-profile"
}
```

### Environment Variables

```bash
export PANW_AI_SEC_API_KEY=your-api-key
export PANW_AI_SEC_PROFILE_NAME=my-security-profile
```

### How It Works

The API key is used to generate an HMAC-SHA256 hash of the request payload, which is sent as the `x-payload-hash` header alongside the API key in `x-pan-token`.

## OAuth2 Authentication

Used for Management, Model Security, and Red Team operations.

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

## Precedence

Provider config attributes always take precedence over environment variables:

1. Explicit provider attribute value
2. Service-specific environment variable (e.g., `PANW_MODEL_SEC_CLIENT_ID`)
3. Fallback environment variable (e.g., `PANW_MGMT_CLIENT_ID`)
