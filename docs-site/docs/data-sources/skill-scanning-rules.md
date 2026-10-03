---
title: Skill Scanning Catalog rules
---

# Skill Scanning Catalog rules

Returns the complete bounded catalog in `result.rules`. Defaults are catalog settings, not effective tenant policy. Pagination continues until an empty page because observed totals can mean page length.

```hcl
data "prisma-airs_supply_chain_skill_scanning_rules" "example" {

}
```

Responses are native Terraform objects and tuples, not JSON strings: access fields directly without `jsondecode`. Optional fields may be omitted or null, following the SDK contract; use `try(...)` where appropriate. Page data sources retain API pagination metadata. Complete rule inventories return their own arrays without inventing global totals.

Discovery traverses offsets until an empty page, with a maximum of 100 pages requesting 100 rows each. The API provides no snapshot or cursor; concurrent service changes can affect this inventory. Run again after changes settle when you need a stable comparison.

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
