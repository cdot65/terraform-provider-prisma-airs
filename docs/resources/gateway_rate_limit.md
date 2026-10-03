---
page_title: "prisma-airs_gateway_rate_limit (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_rate_limit Resource

Manages `prisma-airs_gateway_rate_limit` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_rate_limit" "example" {
  name = "Example - Gateway - Development"
  workspace_id = var.workspace_id
  type = "requests"
  unit = "rpm"
  value = 100
  conditions = [{key = "metadata.application", value = "example"}]
  group_by = [{key = "metadata.application"}]
}
```

## Ownership and lifecycle

Conditions and grouping are native HCL lists of objects. Type, workspace, target, and grouping are immutable. A changed limit or unit updates the resource. An omitted server-defaulted target does not force replacement. Applying Terraform does not generate traffic.

## Import

```bash
terraform import prisma-airs_gateway_rate_limit.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_rate_limit/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `conditions` | `dynamic` | required | — | Native HCL condition objects. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `group_by` | `dynamic` | required | — | Native HCL grouping objects. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Policy name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `target` | `string` | optional, computed | — | Optional policy target. |
| `type` | `string` | required | — | Rate measurement type. |
| `unit` | `string` | required | — | Measurement unit, such as rpm. |
| `value` | `number` | required | — | Rate threshold; zero is explicit. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
