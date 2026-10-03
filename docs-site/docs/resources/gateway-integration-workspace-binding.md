# AI Gateway integration workspace binding

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

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_integration_workspace_binding.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
