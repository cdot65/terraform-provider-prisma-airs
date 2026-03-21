# Prisma AIRS Terraform Provider

Terraform provider for Palo Alto Networks **Prisma AI Runtime Security (AIRS)** — manage AI security infrastructure as code across four service domains.

## Service Domains

| Domain | Resources | Description |
|--------|-----------|-------------|
| **Management** | Profiles, Topics, API Keys, Apps | Security profile and configuration CRUD |
| **Model Security** | Groups, Scans, Rules | ML model scanning and security group management |
| **AI Red Teaming** | Targets, Scans, Custom Attacks | Automated red team testing infrastructure |
| **AI Runtime Security** | Content Scan (data source) | Real-time content scanning |

## Key Features

- **Full AIRS coverage** — resources and data sources for all four service domains
- **Built on prisma-airs-go** — uses the official Go SDK for API interactions
- **Terraform Plugin Framework** — modern provider architecture with typed schemas
- **Dual auth support** — API Key for scans, OAuth2 for management APIs
- **Environment variable fallback** — provider config or env vars for all credentials

## Architecture

```mermaid
graph LR
    A[Terraform] --> B[Provider]
    B --> C[prisma-airs-go SDK]
    C --> D[Scan API<br/>API Key Auth]
    C --> E[Management API<br/>OAuth2]
    C --> F[Model Security API<br/>OAuth2]
    C --> G[Red Team API<br/>OAuth2]
```

## Quick Links

- [Installation](getting-started/installation.md)
- [Quick Start](getting-started/quick-start.md)
- [Configuration](getting-started/configuration.md)
- [Provider Configuration Reference](reference/provider-configuration.md)
