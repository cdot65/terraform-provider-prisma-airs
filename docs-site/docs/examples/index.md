---
title: Examples
slug: /examples
---

Use these complete configurations with the [updated provider](../getting-started/installation.md#use-the-updated-provider-from-source). Examples register and manage configuration; they do not start model scans or red-team inference jobs.

| Example | What it manages | Access |
| --- | --- | --- |
| [Runtime policy](runtime-policy.md) | Custom topic and versioned security profile | Runtime management |
| [Native targets](native-targets.md) | All seven native HCL connection families | Red Team management |
| [Model Security](model-security.md) | Group and rule catalog | Licensed Model Security |
| [Repository configurations](repository-configurations.md) | Multi-resource working examples | Services used by each configuration |

Configure OAuth credentials outside the HCL. Review `terraform plan` and a stable post-apply plan. Cleanup should account for [revision history, archives, tombstones, and key/app cascades](../guides/import-and-state.md).
