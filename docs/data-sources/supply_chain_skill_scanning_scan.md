---
page_title: "prisma-airs_supply_chain_skill_scanning_scan (Data Source)"
subcategory: "AI Supply Chain Security"
---

# prisma-airs_supply_chain_skill_scanning_scan Data Source

# Skill Scanning Scan detail

Provide exactly one of `scan_uuid` or a lowercase SHA-256 `fingerprint`. The latter calls lookup. `result` contains scan status, summaries, policy outcomes, and optional metadata as a sensitive native object. No job is started or polled.

```hcl
data "prisma-airs_supply_chain_skill_scanning_scan" "example" {
  scan_uuid = var.skill_scan_uuid
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

See [Skill Scanning workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/skill-scanning-workflow/).

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `fingerprint` | `string` | optional | — | Lowercase SHA-256 fingerprint for scan lookup. |
| `id` | `string` | computed | — | Discovery identity. |
| `result` | `dynamic` | computed | yes | Native Terraform object containing the SDK response, including nulls and nested lists. Access fields directly (for example result.rules); no JSON decoding is needed. Scans, overrides and findings return one API page. Rules/effective rules use offset traversal until empty (at most 100 pages of 100); concurrent service edits can affect this non-snapshot inventory. |
| `scan_uuid` | `string` | optional | — | Scan UUID. Supply either scan_uuid or fingerprint. |
