---
slug: /overview
page_title: "Prisma AIRS Provider"
---

# Prisma AIRS Terraform Provider

Terraform provider for Palo Alto Networks **Prisma AIRS** — manage AI security infrastructure as code.

## Products

The [product catalog](reference/index.md) derives available resources and data sources from the provider registrations:

- **AI Runtime Security:** profiles, custom topics, API keys, customer apps, DLP profiles, and deployment profiles.
- **AI Red Teaming:** targets and custom prompt sets; the development candidate also adds adapter ownership and discovery.
- **AI Gateway:** managed workspaces and dedicated IAM scopes, routing configs, guardrails, integrations, workspace bindings, MCP servers, keys, limits, secret references, deployments, and metadata discovery.
- **AI Supply Chain Security:** Model Security groups/rules, Skill Scanning tenant configuration, policy rules, trusted skills, and scan discovery.

## Key Features

- **AIRS management coverage** — resources and data sources for the implemented products
- **Built on prisma-airs-go** — uses the project Go SDK for API interactions
- **Terraform Plugin Framework** — modern provider architecture with typed schemas
- **OAuth2 authentication** — client credentials grant for all APIs
- **Environment variable fallback** — provider config or env vars for all credentials

## Architecture

Terraform → provider → prisma-airs-go SDK → AIRS service APIs. See the [architecture guide](development/architecture.md) for the product modules.

These guides describe the current provider, including Skill Scanning support built on Go SDK v0.8.1. See [Getting started](getting-started/index.md) and [Migration](guides/migration.md) for the installation and state transition.

## Quick Links

- [Installation](getting-started/installation.md)
- [Quick Start](getting-started/quick-start.md)
- [Configuration](getting-started/configuration.md)
- [Provider Configuration Reference](reference/provider-configuration.md)

## Example Usage

```hcl
# Setup: Declare the provider required by this configuration.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.11.0"
    }
  }
}

# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {}
```

Configure `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` through your secret manager. See the [provider schema](reference/generated/provider.md) for every argument.

## Product endpoint configuration

OAuth credentials (`client_id`, sensitive `client_secret`, and `tsg_id`) and `token_endpoint` are shared. Endpoint overrides live in optional product blocks; omit them for environment/SDK defaults.

```hcl
# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {
  runtime {
    mgmt_endpoint = "https://api.sase.paloaltonetworks.com/aisec"
  }

  red_team {
    data_endpoint = "https://api.sase.paloaltonetworks.com/ai-red-teaming/data-plane"
    mgmt_endpoint = "https://api.sase.paloaltonetworks.com/ai-red-teaming/mgmt-plane"
  }

  supply_chain {
    data_endpoint = "https://api.sase.paloaltonetworks.com/aims/data"
    mgmt_endpoint = "https://api.sase.paloaltonetworks.com/aims/mgmt"
  }
}
```

Gateway endpoint overrides use the optional `gateway` block. See the [Gateway workflow](guides/gateway-workflow.md). See [product configuration](reference/provider-configuration.md) for environment mappings and [migration](guides/migration.md) before upgrading existing state.
