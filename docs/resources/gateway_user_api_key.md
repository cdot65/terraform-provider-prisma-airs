---
page_title: "prisma-airs_gateway_user_api_key (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_user_api_key Resource

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

An existing `expires_at` remains when the argument is removed. Use a supported explicit future timestamp or deliberately replace the owned object; omission does not disable expiry.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_user_api_key/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_emails` | `set(string)` | optional, computed | — | Usage alert recipients. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `defaults` | `dynamic` | optional, computed | — | Native HCL key defaults; config_id, allow_config_override and metadata. |
| `description` | `string` | optional, computed | — | Key description. |
| `expires_at` | `string` | optional, computed | — | RFC 3339 expiry timestamp; equivalent offsets and precision are preserved on refresh. Removing this setting does not clear a remote expiry. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `key` | `string` | computed | yes | One-time key material; export securely. Import cannot recover it. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Key name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `scopes` | `set(string)` | required | — | Gateway permissions granted to the key. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `user_id` | `string` | required | — | User UUID owning this user key; required by SCM. |
| `workspace_id` | `string` | required | — | Existing Gateway workspace UUID. Workspace and IAM provisioning are external. |
