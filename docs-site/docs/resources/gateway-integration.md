# AI Gateway integration

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

See the [exact schema reference](../reference/generated/prisma-airs_gateway_integration.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
