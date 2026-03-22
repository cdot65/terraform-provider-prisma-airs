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
      version = "~> 0.3"
    }
  }
}

provider "prisma-airs" {}

resource "prisma-airs_security_profile" "example" {
  profile_name = "my-ai-security-profile"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 30
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:allow"
    }
  }
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
  name            = "my-ai-application"
  target_type     = "APPLICATION"
  connection_type = "REST"

  connection_params = jsonencode({
    url = "https://my-app.example.com/api/chat"
    headers = {
      "Content-Type" = "application/json"
    }
    request_json = {
      prompt = "{INPUT}"
    }
    response_json = {
      output = "{RESPONSE}"
    }
    response_key = "output"
  })
}
```

## Next Steps

- [Configuration Reference](configuration.md) — all provider attributes
- [Resources](../resources/security-profile.md) — full resource documentation
- [Guides](../guides/authentication.md) — detailed authentication guide
