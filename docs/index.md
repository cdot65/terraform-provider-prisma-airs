---
page_title: "Prisma AIRS Provider"
---

# Prisma AIRS Terraform Provider

Terraform provider for Palo Alto Networks **Prisma AIRS** — manage AI security infrastructure as code.

## Products

The [product catalog](https://cdot65.github.io/terraform-provider-prisma-airs/reference/) derives available resources and data sources from the provider registrations:

- **AI Runtime Security:** profiles, custom topics, API keys, customer apps, DLP profiles, and deployment profiles.
- **AI Red Teaming:** targets and custom prompt sets.
- **AI Gateway:** routing configs, guardrails, integrations, workspace bindings, MCP servers, keys, limits, secret references, deployments, and metadata discovery.
- **AI Supply Chain Security:** Model Security groups and the rule catalog.

## Key Features

- **AIRS management coverage** — resources and data sources for the implemented products
- **Built on prisma-airs-go** — uses the project Go SDK for API interactions
- **Terraform Plugin Framework** — modern provider architecture with typed schemas
- **OAuth2 authentication** — client credentials grant for all APIs
- **Environment variable fallback** — provider config or env vars for all credentials

## Architecture

Terraform → provider → prisma-airs-go SDK → AIRS service APIs. See the [architecture guide](https://cdot65.github.io/terraform-provider-prisma-airs/development/architecture/) for the product modules.

These guides describe provider v0.9.0 built on Go SDK v0.6.1. See [Getting started](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/) and [Migration](https://cdot65.github.io/terraform-provider-prisma-airs/guides/migration/) for the installation and state transition.

## Quick Links

- [Installation](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/installation/)
- [Quick Start](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/quick-start/)
- [Configuration](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/configuration/)
- [Provider Configuration Reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/provider-configuration/)

## Example Usage

```hcl
# Setup: Declare the provider required by this configuration.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.9.0"
    }
  }
}

# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {}
```

Configure `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` through your secret manager. See the [provider schema](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/provider/) for every argument.

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

Gateway endpoint overrides use the optional `gateway` block. See the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/). See [product configuration](https://cdot65.github.io/terraform-provider-prisma-airs/reference/provider-configuration/) for environment mappings and [migration](https://cdot65.github.io/terraform-provider-prisma-airs/guides/migration/) before upgrading existing state.
