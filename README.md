# terraform-provider-prisma-airs

[![CI](https://github.com/cdot65/terraform-provider-prisma-airs/actions/workflows/ci.yml/badge.svg)](https://github.com/cdot65/terraform-provider-prisma-airs/actions/workflows/ci.yml)
[![Tests](https://github.com/cdot65/terraform-provider-prisma-airs/actions/workflows/test.yml/badge.svg)](https://github.com/cdot65/terraform-provider-prisma-airs/actions/workflows/test.yml)
[![Go 1.25.6+](https://img.shields.io/badge/go-%3E%3D1.25.6-00ADD8)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Terraform provider for Palo Alto Networks **Prisma AIRS** — manage AI security infrastructure as code, organized by product.

Built on the [prisma-airs-go](https://github.com/cdot65/prisma-airs-go) SDK using the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

The latest published provider is v0.9.0. Unreleased Skill Scanning support consumes Go SDK v0.7.0, with native HCL and a Terraform 1.11+ write-only authorization-code input. Review the [migration guide](https://cdot65.github.io/terraform-provider-prisma-airs/guides/migration/) before upgrading existing configurations and state.

## Quick Start

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.9.0"
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

| Product | Current functionality |
| --- | --- |
| AI Runtime Security | Profiles, topics, API keys, customer apps, DLP and deployment catalogs |
| AI Red Teaming | Targets and custom prompt sets |
| AI Gateway | 15 management resources and 13 discovery data sources |
| AI Supply Chain Security | Model Security groups/rules plus Skill Scanning tenant instances, policy rules, trusted skills, and scan discovery |

See the [generated product catalog](https://cdot65.github.io/terraform-provider-prisma-airs/reference/) for exact Terraform names. Endpoint overrides use optional `runtime`, `red_team`, `gateway`, and `supply_chain` blocks; credentials remain shared.

## Documentation

Full docs: **[cdot65.github.io/terraform-provider-prisma-airs](https://cdot65.github.io/terraform-provider-prisma-airs/)**

## Development

```bash
make build          # build provider binary
make check          # fmt + vet + lint + test
make testacc        # acceptance tests (requires credentials)
make docs-serve     # serve Docusaurus locally
```

Documentation tooling requires Node 24+. Run `npm ci --prefix docs-site` once, then `make docs-serve`, `make docs-build`, or `make docs-check`.

## License

[MIT](LICENSE)
