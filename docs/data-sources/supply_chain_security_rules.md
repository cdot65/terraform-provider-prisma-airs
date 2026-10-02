---
page_title: "prisma-airs_supply_chain_security_rules (Data Source)"
subcategory: "AI Supply Chain Security"
---

# prisma-airs_supply_chain_security_rules Data Source

Reads the catalog of model security rules from Prisma AIRS Model Security API. Returns all available rules — no filter parameters are accepted.

## Example Usage

```hcl
data "prisma-airs_supply_chain_security_rules" "all" {}

output "rule_count" {
  value = length(data.prisma-airs_supply_chain_security_rules.all.rules)
}

output "rule_names" {
  value = [for r in data.prisma-airs_supply_chain_security_rules.all.rules : r.name]
}
```

## Argument Reference

This data source takes no arguments.

## Attribute Reference

- `rules` - List of security rules. Each item contains:
    - `uuid` - Rule UUID.
    - `name` - Rule name.
    - `description` - Rule description.
    - `source_type` - Compatible source types (comma-separated).
    - `rule_type` - Rule type (`METADATA`, `ARTIFACT`).
    - `created_at` - Creation timestamp.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_supply_chain_security_rules/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `rules` | `list(object)` | computed | — | List of security rules. |

#### Attributes.rules

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | computed | — | Rule description. |
| `name` | `string` | computed | — | Rule name. |
| `rule_type` | `string` | computed | — | Rule type (METADATA, ARTIFACT). |
| `source_type` | `string` | computed | — | Source type. |
| `uuid` | `string` | computed | — | Rule UUID. |
