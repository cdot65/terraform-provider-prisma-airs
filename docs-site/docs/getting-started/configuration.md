# Configuration

Use one OAuth identity per provider configuration. Credentials are shared across Management, Model Security, and Red Team clients; API endpoint overrides remain service-specific.

## Environment configuration

```hcl
provider "prisma-airs" {}
```

Set `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` through your secret manager. See [Authentication](authentication.md) for setup and permission requirements.

## Explicit configuration

```hcl
variable "airs_client_id" {
  type = string
}

variable "airs_client_secret" {
  type      = string
  sensitive = true
}

variable "airs_tsg_id" {
  type = string
}

provider "prisma-airs" {
  client_id     = var.airs_client_id
  client_secret = var.airs_client_secret
  tsg_id        = var.airs_tsg_id
}
```

Explicit attributes override the corresponding environment variables. The provider accepts a common `token_endpoint` and separate management, Model Security, and Red Team API endpoints. Leave overrides unset to use SDK defaults; the [configuration reference](../reference/provider-configuration.md) lists their exact names.

## Separate tenants

Use provider aliases and explicit credentials for a second tenant:

```hcl
variable "second_client_id" {
  type = string
}

variable "second_client_secret" {
  type      = string
  sensitive = true
}

variable "second_tsg_id" {
  type = string
}

provider "prisma-airs" {
  alias         = "second"
  client_id     = var.second_client_id
  client_secret = var.second_client_secret
  tsg_id        = var.second_tsg_id
}

resource "prisma-airs_model_security_group" "second" {
  provider    = prisma-airs.second
  name        = "second-tenant-models"
  source_type = "HUGGING_FACE"
}
```

Pass aliases explicitly into modules. Verify the second tenant's entitlement before applying Model Security resources.
