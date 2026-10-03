# AI Gateway deployment

Manages `prisma-airs_gateway_deployment` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_deployment" "example" {
  name = "Example - Gateway - Development"
  type = "non_production"
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

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_deployment.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
