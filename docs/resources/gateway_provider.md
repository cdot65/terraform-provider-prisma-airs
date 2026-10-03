---
page_title: "prisma-airs_gateway_provider (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_provider Resource

Manages `prisma-airs_gateway_provider` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_provider" "example" {
  name = "Example - Gateway - Development"
  workspace_id = var.workspace_id
  integration_id = var.integration_id
  note = "Application provider"
}
```

## Ownership and lifecycle

Establish integration workspace access before creation. Add `depends_on` for that binding, since its ID is separate from `integration_id`. Workspace and integration changes require replacement. This resource owns the workspace provider, not the upstream organization integration.

Inline provider rate limits on create currently fail live with upstream HTTP 503 through the published SDK contract. They are excluded from this resource. Manage rate policies with [`prisma-airs_gateway_rate_limit`](https://cdot65.github.io/terraform-provider-prisma-airs/resources/gateway-rate-limit/); they have their own conditions and lifecycle. Provider usage limits expose supported settings as typed HCL.

## Import

```bash
terraform import prisma-airs_gateway_provider.example <workspace_uuid>/<provider_uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_provider/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `expires_at` | `string` | optional, computed | — | RFC 3339 expiry timestamp; equivalent offsets and precision are preserved on refresh. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `integration_id` | `string` | required | — | Bound organisation integration UUID. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Provider name. |
| `note` | `string` | optional, computed | — | Provider note; an empty string clears it. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `usage_limits` | `single(object)` | optional, computed | — | Provider usage limit settings, represented as native HCL attributes. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |

#### Attributes.usage_limits

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_threshold` | `number` | optional, computed | — | Alert threshold; zero is explicit. |
| `credit_limit` | `number` | optional, computed | — | Maximum usage credits. |
| `next_usage_reset_at` | `string` | optional, computed | — | Next reset timestamp. |
| `periodic_reset` | `string` | optional, computed | — | Reset cadence, such as monthly or weekly. |
| `periodic_reset_days` | `number` | optional, computed | — | Custom reset interval in days. |
| `type` | `string` | optional, computed | — | Usage measurement type: cost or tokens. |
