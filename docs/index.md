---
slug: /overview
page_title: "Prisma AIRS Provider"
---

# Prisma AIRS Terraform Provider

Terraform provider for Palo Alto Networks **Prisma AI Runtime Security (AIRS)** — manage AI security infrastructure as code.

## Service Domains

| Domain | Resources | Description |
|--------|-----------|-------------|
| **Management** | Profiles, Topics, API Keys, Apps | Security profile and configuration CRUD |
| **Model Security** | Groups, Rules | Security group management and rules |
| **AI Red Teaming** | Targets, Custom Prompt Sets | Red team target and prompt set management |

## Key Features

- **AIRS management coverage** — resources and data sources for management, model security, and red team domains
- **Built on prisma-airs-go** — uses the project Go SDK for API interactions
- **Terraform Plugin Framework** — modern provider architecture with typed schemas
- **OAuth2 authentication** — client credentials grant for all APIs
- **Environment variable fallback** — provider config or env vars for all credentials

## Architecture

Terraform → provider → prisma-airs-go SDK → AIRS service APIs. See the [architecture guide](development/architecture.md) for the service boundaries.

These guides describe provider v0.7.0 built on Go SDK v0.6.1. See [Getting started](getting-started/index.md) and [Migration](guides/migration.md) for the installation and state transition.

## Quick Links

- [Installation](getting-started/installation.md)
- [Quick Start](getting-started/quick-start.md)
- [Configuration](getting-started/configuration.md)
- [Provider Configuration Reference](reference/provider-configuration.md)

## Example Usage

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.7.0"
    }
  }
}

provider "prisma-airs" {}
```

Configure `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` through your secret manager. See the [provider schema](reference/generated/provider.md) for every argument.
