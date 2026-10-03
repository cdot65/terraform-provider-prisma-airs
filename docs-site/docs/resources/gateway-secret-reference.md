# AI Gateway secret reference

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

Refresh warns when the service omits a configured `allowed_workspaces` list. Terraform retains that desired list, but cannot verify out-of-band changes to it; verify workspace access in Gateway.

## Import

```bash
terraform import prisma-airs_gateway_secret_reference.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_secret_reference.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.
