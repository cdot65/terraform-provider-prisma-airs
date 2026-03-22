# prisma-airs-provider

[![CI](https://github.com/cdot65/prisma-airs-provider/actions/workflows/ci.yml/badge.svg)](https://github.com/cdot65/prisma-airs-provider/actions/workflows/ci.yml)
[![Tests](https://github.com/cdot65/prisma-airs-provider/actions/workflows/test.yml/badge.svg)](https://github.com/cdot65/prisma-airs-provider/actions/workflows/test.yml)
[![Go 1.24+](https://img.shields.io/badge/go-%3E%3D1.24-00ADD8)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Terraform provider for Palo Alto Networks **Prisma AI Runtime Security (AIRS)** — manage AI security infrastructure as code across Management, Model Security, and Red Teaming domains.

Built on the [prisma-airs-go](https://github.com/cdot65/prisma-airs-go) SDK using the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

## Quick Start

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.4"
    }
  }
}

provider "prisma-airs" {}
```

```bash
export PANW_MGMT_CLIENT_ID="your-client-id"
export PANW_MGMT_CLIENT_SECRET="your-client-secret"
export PANW_MGMT_TSG_ID="1234567890"
terraform init && terraform apply
```

## Coverage

| Domain | Resources | Data Sources |
|--------|-----------|--------------|
| Management | `security_profile`, `custom_topic`, `api_key`, `customer_app` | `dlp_profiles`, `deployment_profiles` |
| Model Security | `model_security_group` | `model_security_rules` |
| Red Team | `red_team_target`, `red_team_custom_prompt_set` | — |

## Documentation

Full docs: **[cdot65.github.io/prisma-airs-provider](https://cdot65.github.io/prisma-airs-provider/)**

## Development

```bash
make build          # build provider binary
make check          # fmt + vet + lint + test
make testacc        # acceptance tests (requires credentials)
make docs-serve     # serve mkdocs locally
```

## License

[MIT](LICENSE)
