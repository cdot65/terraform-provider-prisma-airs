---
title: Provider reference
slug: /reference
---

Product ownership is generated from the provider registrations. Lifecycle guides explain imports and updates; exact schemas list types, nested blocks, and sensitive fields.

See [provider configuration](provider-configuration.md), [environment variables](environment-variables.md), and the [provider schema](generated/provider.md).

## AI Runtime Security

| Terraform type | Kind | Guide | Schema |
| --- | --- | --- | --- |
| `prisma-airs_runtime_security_profile` | Resource | [Lifecycle](../resources/security-profile.md) | [Attributes](generated/prisma-airs_runtime_security_profile.md) |
| `prisma-airs_runtime_custom_topic` | Resource | [Lifecycle](../resources/custom-topic.md) | [Attributes](generated/prisma-airs_runtime_custom_topic.md) |
| `prisma-airs_runtime_api_key` | Resource | [Lifecycle](../resources/api-key.md) | [Attributes](generated/prisma-airs_runtime_api_key.md) |
| `prisma-airs_runtime_customer_app` | Resource | [Lifecycle](../resources/customer-app.md) | [Attributes](generated/prisma-airs_runtime_customer_app.md) |
| `prisma-airs_runtime_dlp_profiles` | Data source | [Lifecycle](../data-sources/dlp-profiles.md) | [Attributes](generated/prisma-airs_runtime_dlp_profiles.md) |
| `prisma-airs_runtime_deployment_profiles` | Data source | [Lifecycle](../data-sources/deployment-profiles.md) | [Attributes](generated/prisma-airs_runtime_deployment_profiles.md) |

## AI Red Teaming

| Terraform type | Kind | Guide | Schema |
| --- | --- | --- | --- |
| `prisma-airs_red_team_target` | Resource | [Lifecycle](../resources/red-team-target.md) | [Attributes](generated/prisma-airs_red_team_target.md) |
| `prisma-airs_red_team_custom_prompt_set` | Resource | [Lifecycle](../resources/red-team-custom-prompt-set.md) | [Attributes](generated/prisma-airs_red_team_custom_prompt_set.md) |

## AI Gateway

Not yet implemented. There are no Gateway resources, data sources, or configuration settings in this release. See [AI Gateway status](../products/gateway.md).

## AI Supply Chain Security

| Terraform type | Kind | Guide | Schema |
| --- | --- | --- | --- |
| `prisma-airs_supply_chain_security_group` | Resource | [Lifecycle](../resources/model-security-group.md) | [Attributes](generated/prisma-airs_supply_chain_security_group.md) |
| `prisma-airs_supply_chain_security_rules` | Data source | [Lifecycle](../data-sources/model-security-rules.md) | [Attributes](generated/prisma-airs_supply_chain_security_rules.md) |
