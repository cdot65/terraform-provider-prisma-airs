# AI Gateway provider

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

## Import

```bash
terraform import prisma-airs_gateway_provider.example <workspace_uuid>/<provider_uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_provider.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
