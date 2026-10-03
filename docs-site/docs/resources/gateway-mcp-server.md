# AI Gateway mcp server

Manages `prisma-airs_gateway_mcp_server` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_mcp_server" "example" {
  name = "Example - Gateway - Development"
  workspace_id = var.workspace_id
  mcp_integration_id = var.mcp_integration_id
}
```

## Ownership and lifecycle

Bind the organization MCP integration to this workspace first and declare `depends_on` on the binding. Workspace or MCP integration changes require replacement. Applying Terraform does not run connectivity tests, invoke tools, terminate connections, or configure capabilities/user access.

## Import

```bash
terraform import prisma-airs_gateway_mcp_server.example <workspace_uuid>/<server_uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_mcp_server.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
