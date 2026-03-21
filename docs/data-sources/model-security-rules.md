# prisma-airs_model_security_rules

Reads model security rules from Prisma AIRS Model Security API.

## Example Usage

```hcl
data "prisma-airs_model_security_rules" "all" {}

data "prisma-airs_model_security_rules" "huggingface" {
  source_type = "HUGGING_FACE"
}
```

## Argument Reference

- `limit` - (Optional) Maximum number of rules to return.
- `skip` - (Optional) Number of rules to skip.
- `source_type` - (Optional) Filter by source type.
- `search_query` - (Optional) Search query string.

## Attribute Reference

- `items` - List of security rules. Each item contains:
    - `uuid` - Rule UUID.
    - `name` - Rule name.
    - `description` - Rule description.
    - `source_type` - Source type.
    - `rule_type` - Rule type (`METADATA`, `ARTIFACT`).
    - `created_at` - Timestamp.
- `total` - Total number of rules.
