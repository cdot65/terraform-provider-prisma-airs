# AI Gateway user api key

Manages `prisma-airs_gateway_user_api_key` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_user_api_key" "example" {
  name = "Example - Gateway - Development"
  workspace_id = var.workspace_id
  user_id = var.user_id
  scopes = ["completions.write"]
}
```

## Ownership and lifecycle

SCM requires an explicit owning `user_id`; changing it or the workspace replaces the key. The provider uses the explicit user-key route. The sensitive computed `key` is one-time material retained through masked reads and cannot be recovered by import. Rotation is not automatic; deliberate replacement creates new material and deletes the old owned key.

## Import

```bash
terraform import prisma-airs_gateway_user_api_key.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_user_api_key.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
