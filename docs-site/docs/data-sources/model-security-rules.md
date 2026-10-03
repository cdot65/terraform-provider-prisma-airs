# prisma-airs_supply_chain_security_rules

Reads the catalog of model security rules from Prisma AIRS Model Security API. Returns all available rules — no filter parameters are accepted.

## Example Usage

```hcl
# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_supply_chain_security_rules" "all" {}

# Outputs: Expose results for the next configuration or application step.
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

See the [exact schema reference](../reference/generated/prisma-airs_supply_chain_security_rules.md) for all nested fields, types, and sensitivity flags.
