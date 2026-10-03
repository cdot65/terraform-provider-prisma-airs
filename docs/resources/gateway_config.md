---
page_title: "prisma-airs_gateway_config (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_config Resource

Manages `prisma-airs_gateway_config` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
# Routing: Keep the visible routing document free of upstream credentials.
resource "prisma-airs_gateway_config" "example" {
  name         = "Example - Gateway - Development"
  workspace_id = var.workspace_id

  config = {
    provider = "openai"

    retry = {
      attempts = 1
    }

    cache = {
      mode = "simple"
    }
  }
}
```

## Ownership and lifecycle

`config` accepts a native HCL object or map, including nested lists, explicit nulls, false, zero, and empty collections. JSON strings are rejected. Use provider/integration or secret-reference identifiers; embedded credential fields are rejected because routing documents are visible in plans.

The provider sends the complete desired document on update. A leaf edit remains an update at the same Terraform address and resource ID; `version_id` changes. Remote additions/removals in the document appear as drift. Imported configs containing recognized plaintext credentials are rejected before being stored.

Credential detection includes named/header-style fields and recognizable key/token/PEM values. Model token settings such as `pad_token` and `eos_token` are allowed. Never place any secret in the visible routing document.

Server normalization appears as readable drift on refresh. Apply preserves configured values; provide a document accepted by the service to keep subsequent plans empty.

Legitimate provider settings that resemble credential fields belong in the sensitive integration `configurations` object. Routing validation recognizes common GitHub, Google, Slack and GitLab credential prefixes as well as API-key, bearer and private-key patterns. Model `start_token` and `stop_token` fields remain visible nonsecret settings.

Refresh and import also reject recognized credentials in remote routing documents. Remove the embedded credential remotely and replace it with a reference. If you intend to stop managing that existing config, `terraform state rm prisma-airs_gateway_config.example` relinquishes Terraform ownership while leaving the remote object intact.

## Import

```bash
terraform import prisma-airs_gateway_config.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

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
