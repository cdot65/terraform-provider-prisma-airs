# AI Gateway config

Manages `prisma-airs_gateway_config` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_config" "example" {
  name = "Example - Gateway - Development"
  workspace_id = var.workspace_id
  config = { provider = "openai", retry = { attempts = 1 }, cache = { mode = "simple" } }
}
```

## Ownership and lifecycle

`config` accepts a native HCL object or map, including nested lists, explicit nulls, false, zero, and empty collections. JSON strings are rejected. Use provider/integration or secret-reference identifiers; embedded credential fields are rejected because routing documents are visible in plans.

The provider sends the complete desired document on update. A leaf edit remains an update at the same Terraform address and resource ID; `version_id` changes. Remote additions/removals in the document appear as drift. Imported configs containing recognized plaintext credentials are rejected before being stored.

## Import

```bash
terraform import prisma-airs_gateway_config.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_config.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
