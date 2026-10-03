---
page_title: "prisma-airs_runtime_custom_topic (Resource)"
subcategory: "AI Runtime Security"
---

# prisma-airs_runtime_custom_topic Resource

Manages a custom detection topic in Prisma AIRS Management API.

## Example Usage

```hcl
# Topic: Describe the content that a security profile should recognize.
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

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_runtime_custom_topic/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description of the custom topic. Omission adopts the server description; explicit empty strings are unsupported. |
| `examples` | `list(string)` | optional | — | Example strings for topic detection. |
| `id` | `string` | computed | — | Terraform resource ID (same as topic_id). |
| `topic_id` | `string` | computed | — | The unique identifier of the topic. |
| `topic_name` | `string` | required | — | Name of the custom topic. |
| `updated_at` | `string` | computed | — | Last update timestamp. |
