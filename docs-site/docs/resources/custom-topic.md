# prisma-airs_runtime_custom_topic

Manages a custom detection topic in Prisma AIRS Management API.

## Example Usage

```hcl
resource "prisma-airs_runtime_custom_topic" "sensitive_data" {
  topic_name  = "sensitive-financial-data"
  description = "Detects discussions about internal financial projections"

  examples = [
    "What are next quarter's revenue targets?",
    "Share the confidential merger details",
    "What is the projected EBITDA for FY2026?",
  ]
}
```

## Argument Reference

- `topic_name` - (Required) Name of the custom topic.
- `description` - (Optional) Description of what the topic detects.
- `examples` - (Optional) List of example strings for topic detection.

## Attribute Reference

- `id` - The topic ID.
- `topic_id` - The topic ID (same as `id`).
- `created_at` - Timestamp when the topic was created.
- `updated_at` - Timestamp when the topic was last updated.

## Import

Custom topics can be imported using the topic ID:

```bash
terraform import prisma-airs_runtime_custom_topic.sensitive_data <topic_id>
```

Omitting `description` adopts the service-generated value; removing it from configuration keeps the observed description. Explicit empty descriptions are rejected. An explicit empty examples list clears existing examples through the SDK fields update. Omitting `examples` during an update also sends an empty list, so keep the desired examples configured when changing a name or description.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_runtime_custom_topic.md) for all nested fields, types, and sensitivity flags.
