# AI Gateway usage limit

Manages `prisma-airs_gateway_usage_limit` with provider v0.10.0 and published Go SDK v0.8.1.

## Example

```hcl
# Usage limits: Set a budget; Terraform does not reset accumulated usage.
resource "prisma-airs_gateway_usage_limit" "example" {
  name            = "Example - Gateway - Development"
  workspace_id    = var.workspace_id
  type            = "tokens"
  credit_limit    = 100000
  alert_threshold = 0

  conditions = [
    {
      key   = "metadata.application"
      value = "example"
    }
  ]

  group_by = [
    {
      key = "metadata.application"
    }
  ]
}
```

## Ownership and lifecycle

Conditions and grouping are native HCL lists of objects. Type, workspace, and grouping changes require replacement. `credit_limit` and `alert_threshold` preserve zero. The provider does not reset live usage counters. Omitted optional settings retain server defaults; use explicit supported values when changing them.

## Import

```bash
terraform import prisma-airs_gateway_usage_limit.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_usage_limit.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
