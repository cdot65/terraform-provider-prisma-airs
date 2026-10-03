# prisma-airs_gateway_guardrail schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Gateway guardrail with native HCL checks and actions. Import by UUID. Destroy verifies remote absence.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `actions` | `single(object)` | required | — | Native HCL action object: deny, async, on_success.feedback and on_fail.feedback. |
| `check_parameters` | `dynamic` | optional | yes | Native HCL desired parameter objects keyed by check ID; sensitive and retained through masked reads. |
| `checks` | `list(object)` | required | — | Native HCL check objects with id and optional name and is_enabled. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Guardrail name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `target` | `string` | computed | — | Server-reported guardrail target. |
| `version_id` | `string` | computed | — | Guardrail revision UUID. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |

### Attributes.actions

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `async` | `bool` | required | — | Evaluate asynchronously. |
| `deny` | `bool` | required | — | Deny on failure. |
| `on_fail` | `single(object)` | required | — |  |
| `on_success` | `single(object)` | required | — |  |

#### Attributes.actions.on_fail

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `feedback` | `single(object)` | required | — |  |

##### Attributes.actions.on_fail.feedback

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `metadata` | `string` | required | — | Feedback metadata; an empty string is explicit. |
| `value` | `number` | required | — | Feedback value. |
| `weight` | `number` | required | — | Feedback weight. |

#### Attributes.actions.on_success

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `feedback` | `single(object)` | required | — |  |

##### Attributes.actions.on_success.feedback

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `metadata` | `string` | required | — | Feedback metadata; an empty string is explicit. |
| `value` | `number` | required | — | Feedback value. |
| `weight` | `number` | required | — | Feedback weight. |

### Attributes.checks

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | required | — | Check identifier. |
| `is_enabled` | `bool` | optional, computed | — | Whether the check is enabled; false is explicit. |
| `name` | `string` | optional, computed | — | Optional check name. |
