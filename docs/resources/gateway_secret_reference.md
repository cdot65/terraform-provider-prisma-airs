---
page_title: "prisma-airs_gateway_secret_reference (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_secret_reference Resource

Manages `prisma-airs_gateway_secret_reference` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_secret_reference" "example" {
  name = "Example - Gateway - Development"
  manager_type = "aws_sm"
  auth_config = { aws_auth_type = "serviceRole", aws_region = "us-east-1" }
  secret_path = "application/upstream-api-key"
  allow_all_workspaces = false
  allowed_workspaces = [var.workspace_id]
}
```

## Ownership and lifecycle

This manages a reference to a secret in an external manager. It does not create, read, or delete that external secret. `auth_config` is required, sensitive desired input and is preserved through masked reads; import cannot recover usable credentials. Manager changes replace the reference. The current detail route may omit `allowed_workspaces`; import cannot reconstruct omitted workspace access, so supply it explicitly. Sensitive input removal does not clear remote credentials.

## Import

```bash
terraform import prisma-airs_gateway_secret_reference.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_secret_reference/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `allow_all_workspaces` | `bool` | optional, computed | — | Allow every workspace; false is explicit. |
| `allowed_workspaces` | `set(string)` | optional, computed | — | Explicitly allowed workspace UUIDs; an empty list removes all. |
| `auth_config` | `dynamic` | required | yes | Native HCL external manager authentication settings. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Description. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `manager_type` | `string` | required | — | Secret manager type. |
| `name` | `string` | required | — | Reference name. |
| `organisation_id` | `string` | computed | — | Internal organisation UUID from reads; writes use the shared TSG ID. |
| `secret_key` | `string` | optional, computed | — | Optional key within the upstream secret. |
| `secret_path` | `string` | required | — | External secret path. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `tags` | `dynamic` | optional, computed | — | Native HCL string tag map. |
