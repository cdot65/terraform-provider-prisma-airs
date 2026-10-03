---
title: Skill Scanning Scan and rule statistics
---

# Skill Scanning Scan and rule statistics

Returns `result.scans` and `result.rules`. Periods are `1_HOUR`, `3_HOURS`, `24_HOURS`, `7_DAYS`, `30_DAYS`, and `ALL_TIME`; omission uses the API default. Null means unavailable, not zero. Unique skills count fingerprints; vulnerabilities and rule violations count per scan.

```hcl
data "prisma-airs_supply_chain_skill_scanning_statistics" "example" {
  time_period = "7_DAYS"
}
```

Responses are native Terraform objects and tuples, not JSON strings: access fields directly without `jsondecode`. Optional fields may be omitted or null, following the SDK contract; use `try(...)` where appropriate. Page data sources retain API pagination metadata. Complete rule inventories return their own arrays without inventing global totals.

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

See [Skill Scanning workflow](../guides/skill-scanning-workflow.md).

For example, `try(data.prisma-airs_supply_chain_skill_scanning_statistics.example.result.scans.top_vulnerability, null)` handles an omitted top-vulnerability field; `most_violated_rules` may likewise be omitted or null. This differs from an empty collection or a zero count.
