# Quick Start

## Prerequisites

1. [Install Terraform](https://www.terraform.io/downloads.html) (>= 1.0)
2. Obtain Prisma AIRS credentials (API key for scans, OAuth2 credentials for management)

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
    injection = {
      action = "block"
    }
    toxic_content = {
      action = "alert"
    }
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

resource "prisma-airs_red_team_scan" "security_test" {
  target_uuid = prisma-airs_red_team_target.my_app.id
  job_type    = "STATIC"

  categories = ["SECURITY", "SAFETY"]
}
```

## Example: Content Scanning (Data Source)

```hcl
data "prisma-airs_content_scan" "check" {
  profile_name = "my-security-profile"

  prompt   = "What is the capital of France?"
  response = "The capital of France is Paris."
}

output "scan_verdict" {
  value = data.prisma-airs_content_scan.check.category
}
```

## Next Steps

- [Configuration Reference](configuration.md) — all provider attributes
- [Resources](../resources/security-profile.md) — full resource documentation
- [Guides](../guides/authentication.md) — detailed authentication guide
