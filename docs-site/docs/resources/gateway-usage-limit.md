# AI Gateway usage limit

Manages `prisma-airs_gateway_usage_limit` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_usage_limit" "example" {
  name = "Example - Gateway - Development"
  workspace_id = var.workspace_id
  type = "tokens"
  credit_limit = 100000
  alert_threshold = 0
  conditions = [{key = "metadata.application", value = "example"}]
  group_by = [{key = "metadata.application"}]
}
```

## Ownership and lifecycle

Conditions and grouping are native HCL lists of objects. Type, workspace, and grouping changes require replacement. `credit_limit` and `alert_threshold` preserve zero. The provider does not reset live usage counters. Omitted optional settings retain server defaults; use explicit supported values when changing them.

## Import

```bash
terraform import prisma-airs_gateway_usage_limit.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_usage_limit.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
