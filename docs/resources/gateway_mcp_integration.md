---
page_title: "prisma-airs_gateway_mcp_integration (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_mcp_integration Resource

Manages `prisma-airs_gateway_mcp_integration` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
# MCP connection: Register an external tool service; provision it separately.
resource "prisma-airs_gateway_mcp_integration" "example" {
  name           = "Example - Gateway - Development"
  url            = "https://mcp.deepwiki.com/mcp"
  auth_type      = "none"
  transport      = "http"
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

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_mcp_integration/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `auth_type` | `string` | required | — | MCP authentication type. |
| `configurations` | `dynamic` | optional | yes | Native HCL desired connection settings; arbitrary masked configuration drift cannot be detected. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description; an empty string clears it. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | MCP integration name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `secret_mappings` | `dynamic` | optional, computed | — | Native HCL secret reference mappings. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `transport` | `string` | required | — | MCP transport type. |
| `url` | `string` | required | — | MCP endpoint URL. |
