---
page_title: "prisma-airs_gateway_guardrail (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_guardrail Resource

Manages `prisma-airs_gateway_guardrail` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_guardrail" "example" {
  name = "Example - Gateway - Development"
  workspace_id = var.workspace_id
  checks = [{ id = "default.isAllLowerCase" }]
  actions = {
    deny = false
    async = false
    on_success = { feedback = { value = 5, weight = 1, metadata = "" } }
    on_fail = { feedback = { value = -5, weight = 1, metadata = "" } }
  }
}
```

## Ownership and lifecycle

Checks have typed `id`, `name`, and `is_enabled` fields. Actions have typed booleans and feedback fields. Both use native HCL, serialized internally to the SDK contract. Supply complete nested settings you intend to retain. Workspace changes require replacement.

## Import

```bash
terraform import prisma-airs_gateway_guardrail.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_guardrail/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Sensitive check parameters

Use `check_parameters` for native parameter objects keyed by an ID present in `checks`. The map is sensitive desired input, is retained through masked reads, and cannot be recovered by import. Parameter removal does not guarantee that the service erases an existing credential; set an explicit supported value or replace the guardrail. Check identities and flags remain visible in plans.

## Schema

### Attributes

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

#### Attributes.actions

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `async` | `bool` | required | — | Evaluate asynchronously. |
| `deny` | `bool` | required | — | Deny on failure. |
| `on_fail` | `single(object)` | required | — |  |
| `on_success` | `single(object)` | required | — |  |

##### Attributes.actions.on_fail

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `feedback` | `single(object)` | required | — |  |

###### Attributes.actions.on_fail.feedback

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `metadata` | `string` | required | — | Feedback metadata; an empty string is explicit. |
| `value` | `number` | required | — | Feedback value. |
| `weight` | `number` | required | — | Feedback weight. |

##### Attributes.actions.on_success

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `feedback` | `single(object)` | required | — |  |

###### Attributes.actions.on_success.feedback

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `metadata` | `string` | required | — | Feedback metadata; an empty string is explicit. |
| `value` | `number` | required | — | Feedback value. |
| `weight` | `number` | required | — | Feedback weight. |

#### Attributes.checks

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | required | — | Check identifier. |
| `is_enabled` | `bool` | optional, computed | — | Whether the check is enabled; false is explicit. |
| `name` | `string` | optional, computed | — | Optional check name. |
