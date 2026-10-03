---
page_title: "prisma-airs_gateway_deployment (Resource)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_deployment Resource

Manages `prisma-airs_gateway_deployment` with provider v0.10.0 and published Go SDK v0.8.1.

## Example

```hcl
# Deployment: Register a hybrid gateway; infrastructure installation is external.
resource "prisma-airs_gateway_deployment" "example" {
  name       = "Example - Gateway - Development"
  type       = "non_production"
  is_default = false
}
```

## Ownership and lifecycle

Destroy archives this registration. Archived records may remain in discovery lists; refresh removes an externally archived registration from managed state. `client_auth` and `credentials` are sensitive one-time creation outputs, preserved through masked reads but unavailable on import. Infrastructure provisioning, connection tests, workspace attachment, and authentication rotation remain external. `is_default = true` explicitly changes the tenant default; use false for examples and disposable tests.

## Import

```bash
terraform import prisma-airs_gateway_deployment.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_gateway_deployment/) for every argument, computed value and sensitivity flag, and the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for dependency ordering.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `auth_settings` | `dynamic` | optional | yes | Native HCL desired inbound authentication settings. |
| `client_auth` | `string` | computed | yes | One-time client auth; import cannot recover it. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `credentials` | `dynamic` | computed | yes | One-time deployment credentials; import cannot recover them. |
| `deployment_config` | `dynamic` | optional | yes | Native HCL desired deployment settings; retained through masked reads. |
| `id` | `string` | computed | — | Stable resource identifier. |
| `is_default` | `bool` | required | — | Whether to set this as default; use false for disposable registrations. |
| `last_updated_at` | `string` | computed | — | Last remote update timestamp. |
| `name` | `string` | required | — | Registration name. |
| `slug` | `string` | computed | — | Server-assigned resource slug. |
| `status` | `string` | computed | — | Remote lifecycle status; externally archived objects leave Terraform state. |
| `tags` | `dynamic` | optional, computed | — | Native HCL string tag map. |
| `type` | `string` | required | — | Deployment type, such as non_production. |
