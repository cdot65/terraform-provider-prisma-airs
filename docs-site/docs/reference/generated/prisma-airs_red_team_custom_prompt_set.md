# prisma-airs_red_team_custom_prompt_set schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Red Team custom prompt set.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `active` | `bool` | computed | — | Whether the prompt set is active. |
| `archive` | `bool` | computed | — | Whether the prompt set is archived. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description of the prompt set. Clearing a nonempty value requires replacement; the old set is archived. |
| `id` | `string` | computed | — | Terraform resource ID (same as uuid). |
| `name` | `string` | required | — | Name of the prompt set. |
| `status` | `string` | computed | — | Prompt set status. |
| `updated_at` | `string` | computed | — | Last update timestamp. |
| `uuid` | `string` | computed | — | The unique identifier of the prompt set. |
