# AI Gateway mcp integration

Manages `prisma-airs_gateway_mcp_integration` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_mcp_integration" "example" {
  name = "Example - Gateway - Development"
  url = "https://mcp.deepwiki.com/mcp"
  auth_type = "none"
  transport = "http"
  configurations = {}
}
```

## Ownership and lifecycle

Use the separate MCP workspace binding before creating an MCP server. Sensitive `configurations` are desired inputs retained through masked reads; arbitrary configuration drift cannot be detected. Removing a configured sensitive input does not clear the remote credential; update it explicitly or deliberately replace the integration.

## Import

```bash
terraform import prisma-airs_gateway_mcp_integration.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_mcp_integration.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
