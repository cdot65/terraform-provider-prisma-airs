# prisma-airs_custom_topic schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a custom detection topic.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description of the custom topic. Omission adopts the server description; explicit empty strings are unsupported. |
| `examples` | `list(string)` | optional | — | Example strings for topic detection. |
| `id` | `string` | computed | — | Terraform resource ID (same as topic_id). |
| `topic_id` | `string` | computed | — | The unique identifier of the topic. |
| `topic_name` | `string` | required | — | Name of the custom topic. |
| `updated_at` | `string` | computed | — | Last update timestamp. |
