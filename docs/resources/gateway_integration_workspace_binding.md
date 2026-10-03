---
page_title: "prisma-airs_gateway_integration_workspace_binding (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_integration_workspace_binding Resource

Manages `prisma-airs_gateway_integration_workspace_binding` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_integration_workspace_binding" "example" {
  integration_id = var.integration_id
  workspace_id = var.workspace_id
}
```

## Ownership and lifecycle

This owns one enabled organization integration/workspace pair. It merges that pair without overriding other workspace access and never creates a default provider. An already enabled pair must be imported. Destroy disables only this pair. Remove its workspace providers first using Terraform dependencies; the workspace and integration remain.

## Import

```bash
terraform import prisma-airs_gateway_integration_workspace_binding.example <integration_uuid>/<workspace_uuid>
```

Import requires an existing enabled pair and an active parent integration. This resource has no secret inputs or outputs. It owns explicit workspace access only: global access settings are left unchanged, so disabling this pair does not revoke access granted by a global policy.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_integration_workspace_binding/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Stable resource identifier. |
| `integration_id` | `string` | required | — | Organisation integration UUID. |
| `workspace_id` | `string` | required | — | Existing workspace UUID. |
