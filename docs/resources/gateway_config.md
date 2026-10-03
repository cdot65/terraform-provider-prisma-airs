---
page_title: "prisma-airs_gateway_config (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_config Resource

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

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_config/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `config` | `dynamic` | required | — | Complete native HCL routing document, without plaintext credentials. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Config name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `version_id` | `string` | computed | — | Current immutable config revision UUID; changes on update. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
