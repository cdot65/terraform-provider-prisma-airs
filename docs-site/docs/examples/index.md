---
title: Examples
slug: /examples
---

Choose a product, load its credentials from your secret store, and follow a configuration through plan, apply, use, and cleanup. Start with a complete project when you want a working directory; use the focused configurations below to learn an individual provider concept.

## Complete product projects

The [public examples repository](https://github.com/cdot65/prisma-airs-terraform-examples) contains independent Terraform roots with their own getting-started guides and sanitized live results.

| Product | What you configure |
| --- | --- |
| [AI Runtime Security](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-runtime-security) | A custom topic and application security profile |
| [AI Red Teaming](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-red-teaming) | An authenticated application target and prompt-set container |
| [AI Gateway](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-gateway) | Owned model connections, AIRS checks, four routing patterns, application keys, policies, and optional platform features |
| [AI Supply Chain Security](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-supply-chain-security) | A Model Security group and rule discovery |

Each project has its own prerequisites and state. Apply the product you need; applying all four is unnecessary.

## Focused configurations

These complete inline configurations use the [released provider](../getting-started/installation.md). They register and manage configuration; apply does not start model scans, Red Team assessments, or Gateway inference requests.

| Example | What it teaches | Access |
| --- | --- | --- |
| [Runtime policy](runtime-policy.md) | Topic references and versioned security profiles | Runtime management |
| [Native targets](native-targets.md) | Seven native connection families and payload templates | Red Team management |
| [Gateway](gateway.md) | Integration → workspace binding → provider → routing config | Gateway management and an existing workspace |
| [Model Security](model-security.md) | Groups and read-only rule discovery | Licensed Model Security |
| [Repository configurations](repository-configurations.md) | Complete projects and additional provider-checkout scenarios | Services used by each configuration |

## Read the configuration

Examples use short `# Concept: purpose` comments at dependency and product boundaries. Nested objects and lists are expanded so request shapes, references, and sensitive inputs are easy to follow. Variable descriptions and surrounding guide text explain individual inputs.

Load OAuth credentials outside the HCL. Review a saved plan before applying, then check for a stable post-apply plan. Cleanup accounts for [revision history, archives, tombstones, and key/app cascades](../guides/import-and-state.md).

[Gateway workspace and routing](gateway-workspaces.md) creates a dedicated scope, workspace and child config.
