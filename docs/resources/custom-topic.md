# prisma-airs_custom_topic

Manages a custom detection topic in Prisma AIRS Management API.

## Example Usage

```hcl
resource "prisma-airs_custom_topic" "sensitive_data" {
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
terraform import prisma-airs_custom_topic.sensitive_data <topic_id>
```
