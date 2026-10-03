# AI Gateway provider

Manages `prisma-airs_gateway_provider` with provider v0.10.0 and published Go SDK v0.8.1.

## Example

```hcl
# Provider: Expose the integration after its workspace binding exists.
resource "prisma-airs_gateway_provider" "example" {
  name           = "Example - Gateway - Development"
  workspace_id   = var.workspace_id
  integration_id = var.integration_id
  note           = "Application provider"
}
```

## Ownership and lifecycle

Establish integration workspace access before creation. Add `depends_on` for that binding, since its ID is separate from `integration_id`. Workspace and integration changes require replacement. This resource owns the workspace provider, not the upstream organization integration.

Inline provider rate limits on create currently fail live with upstream HTTP 503 through the published SDK contract. They are excluded from this resource. Manage rate policies with [`prisma-airs_gateway_rate_limit`](gateway-rate-limit.md); they have their own conditions and lifecycle. Provider usage limits expose supported settings as typed HCL.

## Import

```bash
terraform import prisma-airs_gateway_provider.example <workspace_uuid>/<provider_uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

An existing `expires_at` remains when the argument is removed. Use a supported explicit future timestamp or deliberately replace the owned object; omission does not disable expiry.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_provider.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
