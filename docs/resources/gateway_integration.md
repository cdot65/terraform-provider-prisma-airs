---
page_title: "prisma-airs_gateway_integration (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_integration Resource

Manages `prisma-airs_gateway_integration` with provider v0.10.0 and published Go SDK v0.8.1.

## Example

```hcl
# Connection: Keep upstream credentials in the integration, outside routing.
resource "prisma-airs_gateway_integration" "example" {
  name           = "Example - Gateway - Development"
  ai_provider_id = var.ai_provider_id
  key            = var.upstream_api_key
  description    = "Application integration"
}
```

## Ownership and lifecycle

The resource owns an organization integration. It does not create a default provider or bind workspaces implicitly. Establish an `integration_workspace_binding`, then create a workspace provider with `depends_on` on that binding. `ai_provider_id` is immutable. Use `airs cli aigateway integrations providers` to discover the provider family UUID. Sensitive `key` and `configurations` are desired inputs retained through masked reads; their remote drift cannot be recovered reliably.

## Import

```bash
terraform import prisma-airs_gateway_integration.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_integration/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `ai_provider_id` | `string` | required | — | AI provider family UUID from the Gateway catalog. |
| `configurations` | `dynamic` | optional | yes | Native HCL provider settings; masked reads cannot detect arbitrary credential/configuration drift. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Integration description; an empty string clears it. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `key` | `string` | optional | yes | Desired upstream provider credential; not recoverable on import. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Integration name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `pricing_adjustments` | `dynamic` | optional, computed | — | Native HCL pricing adjustments. |
| `secret_mappings` | `dynamic` | optional, computed | — | Native HCL mappings: target_field, secret_reference_id and optional secret_key/value_format. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
