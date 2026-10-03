---
page_title: "prisma-airs_supply_chain_skill_scanning_scans (Data Source)"
subcategory: "AI Supply Chain Security"
---

# prisma-airs_supply_chain_skill_scanning_scans Data Source

# Skill Scanning Scan inventory

Returns one page in `result.scans`. Filters include repeated statuses/artifact types, search, fingerprint, creation-time order, and RFC3339 time bounds. Parent/batch scans and their children must not be double-counted. This result is sensitive.

```hcl
data "prisma-airs_supply_chain_skill_scanning_scans" "example" {
  limit = 10
  statuses = ["COMPLETED", "FAILED"]
  artifact_types = ["SKILL"]
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
| `artifact_types` | `list(string)` | optional | — | Repeated artifact type filters. |
| `end_time` | `string` | optional | — | RFC3339 end time. |
| `fingerprint` | `string` | optional | — | Exact fingerprint filter. |
| `id` | `string` | computed | — | Discovery identity. |
| `limit` | `number` | optional | — | Page size from 1 to 100; omission uses the server default. |
| `result` | `dynamic` | computed | yes | Native Terraform object containing the SDK response, including nulls and nested lists. Access fields directly (for example result.rules); no JSON decoding is needed. Scans, overrides and findings return one API page. Rules/effective rules use offset traversal until empty (at most 100 pages of 100); concurrent service edits can affect this non-snapshot inventory. |
| `search_query` | `string` | optional | — | Search by scan name or UUID. |
| `skip` | `number` | optional | — | Nonnegative page offset. |
| `sort_order` | `string` | optional | — | Creation-time sort order. |
| `start_time` | `string` | optional | — | RFC3339 start time. |
| `statuses` | `list(string)` | optional | — | Repeated status filters. |
