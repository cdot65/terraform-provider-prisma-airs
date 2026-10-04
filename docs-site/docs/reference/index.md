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

| Terraform type | Kind | Guide | Schema |
| --- | --- | --- | --- |
| `prisma-airs_gateway_config` | Resource | [Lifecycle](../resources/gateway-config.md) | [Attributes](generated/prisma-airs_gateway_config.md) |
| `prisma-airs_gateway_guardrail` | Resource | [Lifecycle](../resources/gateway-guardrail.md) | [Attributes](generated/prisma-airs_gateway_guardrail.md) |
| `prisma-airs_gateway_org_guardrail` | Resource | [Lifecycle](../resources/gateway-org-guardrail.md) | [Attributes](generated/prisma-airs_gateway_org_guardrail.md) |
| `prisma-airs_gateway_integration` | Resource | [Lifecycle](../resources/gateway-integration.md) | [Attributes](generated/prisma-airs_gateway_integration.md) |
| `prisma-airs_gateway_provider` | Resource | [Lifecycle](../resources/gateway-provider.md) | [Attributes](generated/prisma-airs_gateway_provider.md) |
| `prisma-airs_gateway_mcp_integration` | Resource | [Lifecycle](../resources/gateway-mcp-integration.md) | [Attributes](generated/prisma-airs_gateway_mcp_integration.md) |
| `prisma-airs_gateway_mcp_server` | Resource | [Lifecycle](../resources/gateway-mcp-server.md) | [Attributes](generated/prisma-airs_gateway_mcp_server.md) |
| `prisma-airs_gateway_service_api_key` | Resource | [Lifecycle](../resources/gateway-service-api-key.md) | [Attributes](generated/prisma-airs_gateway_service_api_key.md) |
| `prisma-airs_gateway_user_api_key` | Resource | [Lifecycle](../resources/gateway-user-api-key.md) | [Attributes](generated/prisma-airs_gateway_user_api_key.md) |
| `prisma-airs_gateway_usage_limit` | Resource | [Lifecycle](../resources/gateway-usage-limit.md) | [Attributes](generated/prisma-airs_gateway_usage_limit.md) |
| `prisma-airs_gateway_rate_limit` | Resource | [Lifecycle](../resources/gateway-rate-limit.md) | [Attributes](generated/prisma-airs_gateway_rate_limit.md) |
| `prisma-airs_gateway_secret_reference` | Resource | [Lifecycle](../resources/gateway-secret-reference.md) | [Attributes](generated/prisma-airs_gateway_secret_reference.md) |
| `prisma-airs_gateway_deployment` | Resource | [Lifecycle](../resources/gateway-deployment.md) | [Attributes](generated/prisma-airs_gateway_deployment.md) |
| `prisma-airs_gateway_integration_workspace_binding` | Resource | [Lifecycle](../resources/gateway-integration-workspace-binding.md) | [Attributes](generated/prisma-airs_gateway_integration_workspace_binding.md) |
| `prisma-airs_gateway_mcp_integration_workspace_binding` | Resource | [Lifecycle](../resources/gateway-mcp-integration-workspace-binding.md) | [Attributes](generated/prisma-airs_gateway_mcp_integration_workspace_binding.md) |
| `prisma-airs_gateway_workspace` | Resource | [Lifecycle](../resources/gateway-workspace.md) | [Attributes](generated/prisma-airs_gateway_workspace.md) |
| `prisma-airs_gateway_configs` | Data source | [Lifecycle](../data-sources/gateway-configs.md) | [Attributes](generated/prisma-airs_gateway_configs.md) |
| `prisma-airs_gateway_guardrails` | Data source | [Lifecycle](../data-sources/gateway-guardrails.md) | [Attributes](generated/prisma-airs_gateway_guardrails.md) |
| `prisma-airs_gateway_org_guardrails` | Data source | [Lifecycle](../data-sources/gateway-org-guardrails.md) | [Attributes](generated/prisma-airs_gateway_org_guardrails.md) |
| `prisma-airs_gateway_integrations` | Data source | [Lifecycle](../data-sources/gateway-integrations.md) | [Attributes](generated/prisma-airs_gateway_integrations.md) |
| `prisma-airs_gateway_providers` | Data source | [Lifecycle](../data-sources/gateway-providers.md) | [Attributes](generated/prisma-airs_gateway_providers.md) |
| `prisma-airs_gateway_mcp_integrations` | Data source | [Lifecycle](../data-sources/gateway-mcp-integrations.md) | [Attributes](generated/prisma-airs_gateway_mcp_integrations.md) |
| `prisma-airs_gateway_mcp_servers` | Data source | [Lifecycle](../data-sources/gateway-mcp-servers.md) | [Attributes](generated/prisma-airs_gateway_mcp_servers.md) |
| `prisma-airs_gateway_service_api_keys` | Data source | [Lifecycle](../data-sources/gateway-service-api-keys.md) | [Attributes](generated/prisma-airs_gateway_service_api_keys.md) |
| `prisma-airs_gateway_user_api_keys` | Data source | [Lifecycle](../data-sources/gateway-user-api-keys.md) | [Attributes](generated/prisma-airs_gateway_user_api_keys.md) |
| `prisma-airs_gateway_usage_limits` | Data source | [Lifecycle](../data-sources/gateway-usage-limits.md) | [Attributes](generated/prisma-airs_gateway_usage_limits.md) |
| `prisma-airs_gateway_rate_limits` | Data source | [Lifecycle](../data-sources/gateway-rate-limits.md) | [Attributes](generated/prisma-airs_gateway_rate_limits.md) |
| `prisma-airs_gateway_secret_references` | Data source | [Lifecycle](../data-sources/gateway-secret-references.md) | [Attributes](generated/prisma-airs_gateway_secret_references.md) |
| `prisma-airs_gateway_deployments` | Data source | [Lifecycle](../data-sources/gateway-deployments.md) | [Attributes](generated/prisma-airs_gateway_deployments.md) |
| `prisma-airs_gateway_workspace` | Data source | [Lifecycle](../data-sources/gateway-workspace.md) | [Attributes](generated/data-source-prisma-airs_gateway_workspace.md) |
| `prisma-airs_gateway_workspaces` | Data source | [Lifecycle](../data-sources/gateway-workspaces.md) | [Attributes](generated/prisma-airs_gateway_workspaces.md) |
| `prisma-airs_gateway_ai_providers` | Data source | [Lifecycle](../data-sources/gateway-ai-providers.md) | [Attributes](generated/prisma-airs_gateway_ai_providers.md) |

## AI Supply Chain Security

| Terraform type | Kind | Guide | Schema |
| --- | --- | --- | --- |
| `prisma-airs_supply_chain_security_group` | Resource | [Lifecycle](../resources/model-security-group.md) | [Attributes](generated/prisma-airs_supply_chain_security_group.md) |
| `prisma-airs_supply_chain_skill_scanning_instance` | Resource | [Lifecycle](../resources/skill-scanning-instance.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_instance.md) |
| `prisma-airs_supply_chain_skill_scanning_rule` | Resource | [Lifecycle](../resources/skill-scanning-rule.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_rule.md) |
| `prisma-airs_supply_chain_skill_scanning_override` | Resource | [Lifecycle](../resources/skill-scanning-override.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_override.md) |
| `prisma-airs_supply_chain_security_rules` | Data source | [Lifecycle](../data-sources/model-security-rules.md) | [Attributes](generated/prisma-airs_supply_chain_security_rules.md) |
| `prisma-airs_supply_chain_skill_scanning_instance` | Data source | [Lifecycle](../data-sources/skill-scanning-instance.md) | [Attributes](generated/data-source-prisma-airs_supply_chain_skill_scanning_instance.md) |
| `prisma-airs_supply_chain_skill_scanning_rules` | Data source | [Lifecycle](../data-sources/skill-scanning-rules.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_rules.md) |
| `prisma-airs_supply_chain_skill_scanning_rule_instances` | Data source | [Lifecycle](../data-sources/skill-scanning-rule-instances.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_rule_instances.md) |
| `prisma-airs_supply_chain_skill_scanning_overrides` | Data source | [Lifecycle](../data-sources/skill-scanning-overrides.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_overrides.md) |
| `prisma-airs_supply_chain_skill_scanning_scan` | Data source | [Lifecycle](../data-sources/skill-scanning-scan.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_scan.md) |
| `prisma-airs_supply_chain_skill_scanning_scans` | Data source | [Lifecycle](../data-sources/skill-scanning-scans.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_scans.md) |
| `prisma-airs_supply_chain_skill_scanning_vulnerabilities` | Data source | [Lifecycle](../data-sources/skill-scanning-vulnerabilities.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_vulnerabilities.md) |
| `prisma-airs_supply_chain_skill_scanning_attack_chains` | Data source | [Lifecycle](../data-sources/skill-scanning-attack-chains.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_attack_chains.md) |
| `prisma-airs_supply_chain_skill_scanning_statistics` | Data source | [Lifecycle](../data-sources/skill-scanning-statistics.md) | [Attributes](generated/prisma-airs_supply_chain_skill_scanning_statistics.md) |
