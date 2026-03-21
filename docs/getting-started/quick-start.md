# Quick Start

## Prerequisites

1. [Install Terraform](https://www.terraform.io/downloads.html) (>= 1.0)
2. Obtain Prisma AIRS OAuth2 credentials (client ID, client secret, TSG ID)

## Example: Manage a Security Profile

### 1. Set environment variables

```bash
export PANW_MGMT_CLIENT_ID=your-client-id
export PANW_MGMT_CLIENT_SECRET=your-client-secret
export PANW_MGMT_TSG_ID=1234567890
```

### 2. Create Terraform configuration

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.1"
    }
  }
}

provider "prisma-airs" {}

resource "prisma-airs_security_profile" "example" {
  profile_name = "my-ai-security-profile"

  policy = jsonencode({
    ai-security-profiles = [{
      latency-config = {
        inline-timeout-action = "allow"
        max-inline-latency    = 5000
      }
      model-protection = {
        prompt-injection = {
          action = "block"
        }
        toxic-content = {
          action        = "alert"
          toxic-category-list = [
            { category = "profanity", threshold = "low" }
          ]
        }
        url-filtering = {
          action = "alert"
        }
      }
    }]
  })
}
```

### 3. Apply

```bash
terraform init
terraform plan
terraform apply
```

## Example: Create a Red Team Target

```hcl
resource "prisma-airs_red_team_target" "my_app" {
  name        = "my-ai-application"
  target_type = "APPLICATION"

  connection = {
    type     = "REST"
    endpoint = "https://my-app.example.com/api/chat"
  }
}
```

## Next Steps

- [Configuration Reference](configuration.md) — all provider attributes
- [Resources](../resources/security-profile.md) — full resource documentation
- [Guides](../guides/authentication.md) — detailed authentication guide
