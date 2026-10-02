---
title: Provider reference
slug: /reference
---

The updated provider exposes seven resources and three data sources. Resource guides explain lifecycle and imports; schema pages list the exact attributes, nested blocks, types, and sensitive fields.

## Resources

| Resource | Guide | Schema |
| --- | --- | --- |
| Security profile | [Revision ownership](../resources/security-profile.md) | [Attributes](generated/prisma-airs_security_profile.md) |
| Custom topic | [Topics](../resources/custom-topic.md) | [Attributes](generated/prisma-airs_custom_topic.md) |
| API key | [Keys](../resources/api-key.md) | [Attributes](generated/prisma-airs_api_key.md) |
| Customer app | [Import-only apps](../resources/customer-app.md) | [Attributes](generated/prisma-airs_customer_app.md) |
| Model Security group | [Groups](../resources/model-security-group.md) | [Attributes](generated/prisma-airs_model_security_group.md) |
| Red Team target | [Native connections](../resources/red-team-target.md) | [Attributes](generated/prisma-airs_red_team_target.md) |
| Custom prompt set | [Prompt sets](../resources/red-team-custom-prompt-set.md) | [Attributes](generated/prisma-airs_red_team_custom_prompt_set.md) |

## Data sources and configuration

| Read | Guide | Schema |
| --- | --- | --- |
| DLP profiles | [DLP profiles](../data-sources/dlp-profiles.md) | [Attributes](generated/prisma-airs_dlp_profiles.md) |
| Deployment profiles | [Sensitive deployment details](../data-sources/deployment-profiles.md) | [Attributes](generated/prisma-airs_deployment_profiles.md) |
| Model Security rules | [Rule catalog](../data-sources/model-security-rules.md) | [Attributes](generated/prisma-airs_model_security_rules.md) |
| Provider | [Configuration](provider-configuration.md) | [Attributes](generated/provider.md) |

See [environment variables](environment-variables.md) for supported fallbacks and [error handling](error-handling.md) for diagnostics.
