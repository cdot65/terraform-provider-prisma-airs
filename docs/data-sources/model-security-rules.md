# prisma-airs_model_security_rules

Reads the catalog of model security rules from Prisma AIRS Model Security API. Returns all available rules — no filter parameters are accepted.

## Example Usage

```hcl
data "prisma-airs_model_security_rules" "all" {}

output "rule_count" {
  value = length(data.prisma-airs_model_security_rules.all.rules)
}

output "rule_names" {
  value = [for r in data.prisma-airs_model_security_rules.all.rules : r.name]
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
