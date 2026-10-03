---
title: Skill Scanning workflow
---

# Manage Skill Scanning configuration

Skill Scanning is part of AI Supply Chain Security. The provider consumes Go SDK v0.8.1's preview contracts. It manages persistent tenant configuration and reads existing scan results; upload reservation, archive transfer, analysis submission, polling, and CSV export remain SDK/CLI operations.

1. Configure shared OAuth credentials and both Skill Scanning base URLs.
2. Discover catalog rules and effective policy through the rules and rule-instances data sources.
3. Manage each catalog rule UUID once, with `prisma-airs_supply_chain_skill_scanning_rule`.
4. Add reviewed fingerprint trust through `prisma-airs_supply_chain_skill_scanning_override`.
5. Import existing tenant instances before managing their onboarding configuration.

| Terraform resource | Create | Update | Destroy | Import |
| --- | --- | --- | --- | --- |
| `supply_chain_skill_scanning_instance` | Provision absent tenant instance | PUT tenant settings | Delete instance; verify absence | Tenant ID |
| `supply_chain_skill_scanning_rule` | Capture baseline; PUT one rule | PUT one rule, stable ID | Restore captured baseline; verify effective state | Catalog rule UUID |
| `supply_chain_skill_scanning_override` | Trust fingerprint | Replacement (API has no update) | Delete trust; verify absence | Override UUID |

All names have the `prisma-airs_` prefix. Native HCL inputs expose field changes; no resource accepts a JSON document. Instance authorization codes require Terraform 1.11+, are read from configuration, and never enter plan or state. Increment `auth_code_version` when changing or clearing the code. IAM/registration metadata retains desired state because GET may omit or normalize it. Read-only results are native objects; sensitive scan, finding, graph, instance, and trust results are marked sensitive.

Rules are shared tenant policy. Separate resources update only their own rule, but multiple workspaces must not own the same tenant/rule pair. Destroy restores computed `original_state`, recorded at adoption; importing later establishes a new baseline. Catalog rule deletion and individual scan deletion do not exist. A catalog-default baseline is restored as an explicit effective state; the API cannot restore the absence of an instance.

Successful creates retain an identity before follow-up reads. If a failed read leaves a resource tainted, fix access and inspect the retained identity before retrying; import or untaint the confirmed object to avoid duplicate creation. Read failures never erase state unless absence is established. Destroy confirms deletion/restoration independently.

See the [complete HCL example](../examples/skill-scanning.md), [Supply Chain catalog](../reference/index.md), and [verification evidence](../development/skill-scanning-verification.md).

Skill Scanning belongs to **AI Supply Chain Security**. Configure both base URLs in `supply_chain`, using shared OAuth credentials. The Go SDK package name remains `aisec/agentguard`; Terraform names and guides use Skill Scanning.

```hcl
provider "prisma-airs" {
  supply_chain {
    skill_scanning_data_endpoint = "https://api.apps.paloaltonetworks.com/aiag/data"
    skill_scanning_mgmt_endpoint = "https://api.apps.paloaltonetworks.com/aiag/mgmt"
  }
}
```

Alternatively set `PANW_SKILL_SCANNING_DATA_ENDPOINT` and `PANW_SKILL_SCANNING_MGMT_ENDPOINT`. The established SDK variables `PANW_AGENT_GUARD_*_ENDPOINT` are fallback aliases. No endpoint is inferred from credentials.
