# prisma-airs-provider

[![CI](https://github.com/cdot65/prisma-airs-provider/actions/workflows/ci.yml/badge.svg)](https://github.com/cdot65/prisma-airs-provider/actions/workflows/ci.yml)
[![Tests](https://github.com/cdot65/prisma-airs-provider/actions/workflows/test.yml/badge.svg)](https://github.com/cdot65/prisma-airs-provider/actions/workflows/test.yml)
[![Go 1.22+](https://img.shields.io/badge/go-%3E%3D1.22-00ADD8)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Terraform provider for Palo Alto Networks **Prisma AIRS** — managing AI security infrastructure across all four service domains: **AI Runtime Security**, **Management**, **Model Security**, and **AI Red Teaming**.

Built on the [prisma-airs-go](https://github.com/cdot65/prisma-airs-go) SDK.

## Installation

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.1"
    }
  }
}
```

## Provider Configuration

```hcl
provider "prisma-airs" {
  # AI Runtime Security (API Key auth)
  api_key      = var.panw_api_key       # or PANW_AI_SEC_API_KEY
  profile_name = var.panw_profile_name  # or PANW_AI_SEC_PROFILE_NAME

  # Management / Model Security / Red Team (OAuth2 auth)
  client_id     = var.panw_client_id     # or PANW_MGMT_CLIENT_ID
  client_secret = var.panw_client_secret # or PANW_MGMT_CLIENT_SECRET
  tsg_id        = var.panw_tsg_id        # or PANW_MGMT_TSG_ID
}
```

## Resources & Data Sources

### Management API

| Type | Name | Description |
|------|------|-------------|
| Resource | `prisma-airs_security_profile` | AI security profile CRUD |
| Resource | `prisma-airs_custom_topic` | Custom detection topic CRUD |
| Resource | `prisma-airs_api_key` | API key lifecycle management |
| Resource | `prisma-airs_customer_app` | Customer application management |
| Data Source | `prisma-airs_dlp_profiles` | DLP data profile listing |
| Data Source | `prisma-airs_deployment_profiles` | Deployment profile listing |
| Data Source | `prisma-airs_scan_logs` | Scan activity log queries |

### Model Security API

| Type | Name | Description |
|------|------|-------------|
| Resource | `prisma-airs_model_security_group` | Security group CRUD |
| Resource | `prisma-airs_model_scan` | ML model scan management |
| Data Source | `prisma-airs_model_security_rules` | Security rule listing |
| Data Source | `prisma-airs_model_scan_evaluations` | Scan evaluation results |
| Data Source | `prisma-airs_model_scan_violations` | Scan violation results |

### Red Team API

| Type | Name | Description |
|------|------|-------------|
| Resource | `prisma-airs_red_team_target` | Red team target CRUD |
| Resource | `prisma-airs_red_team_scan` | Red team scan management |
| Resource | `prisma-airs_red_team_custom_prompt_set` | Custom prompt set management |
| Data Source | `prisma-airs_red_team_reports` | Scan report data |
| Data Source | `prisma-airs_red_team_categories` | Attack category listing |
| Data Source | `prisma-airs_red_team_quota` | Quota information |

### Scan API

| Type | Name | Description |
|------|------|-------------|
| Data Source | `prisma-airs_content_scan` | Synchronous content scanning |

## Authentication

| Auth Method | Used By |
|-------------|---------|
| **API Key** (HMAC-SHA256) | AI Runtime Security scans only |
| **OAuth2** (client_credentials) | Management, Model Security, Red Team |

```bash
# AI Runtime Security scans
export PANW_AI_SEC_API_KEY=your-api-key

# OAuth2 (shared by Management, Red Team, Model Security)
export PANW_MGMT_CLIENT_ID=your-client-id
export PANW_MGMT_CLIENT_SECRET=your-client-secret
export PANW_MGMT_TSG_ID=1234567890
```

## Documentation

Full documentation at **[cdot65.github.io/prisma-airs-provider](https://cdot65.github.io/prisma-airs-provider/)** — includes resource/data source guides, configuration reference, and examples.

## Development

```bash
make build          # build provider binary
make test           # go test -race ./...
make testacc        # acceptance tests (requires credentials)
make test-coverage  # coverage report
make lint           # golangci-lint
make check          # fmt + vet + lint + test
make install        # install locally for development
make generate       # generate provider docs
make docs-serve     # serve mkdocs locally
```

## License

MIT
