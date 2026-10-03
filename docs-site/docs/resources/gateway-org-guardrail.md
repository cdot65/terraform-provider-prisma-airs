# AI Gateway org guardrail

Manages `prisma-airs_gateway_org_guardrail` with provider v0.9.0 and published Go SDK v0.6.1.

## Example

```hcl
resource "prisma-airs_gateway_org_guardrail" "example" {
  name = "Example - Gateway - Development"
  checks = [{ id = "default.isAllLowerCase" }]
  actions = {
    deny = false
    async = false
    on_success = { feedback = { value = 5, weight = 1, metadata = "" } }
    on_fail = { feedback = { value = -5, weight = 1, metadata = "" } }
  }
}
```

## Ownership and lifecycle

This uses the admin-plane organization route. OAuth TSG defaults select the organization; no workspace argument is accepted. Supply complete native checks and actions.

## Import

```bash
terraform import prisma-airs_gateway_org_guardrail.example <uuid>
```

Import reads the active remote object. Archived objects are rejected. Sensitive desired inputs and one-time outputs are not recovered from masked reads; supply desired inputs and preserve previously exported secrets externally.

## Optional values

Omitting an optional setting leaves the remote value or service default in place; removing it from HCL does not clear it. Use an explicit empty string for descriptions/notes documented as clearable. Collection, object and timestamp removal does not send a remote reset; retain an explicit supported value or replace the owned resource deliberately. Sensitive settings remain desired inputs through masked reads.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_gateway_org_guardrail.md) for every argument, computed value and sensitivity flag, and the [Gateway workflow](../guides/gateway-workflow.md) for dependency ordering.

## Sensitive check parameters

Use `check_parameters` for native parameter objects keyed by an ID present in `checks`. The map is sensitive desired input, is retained through masked reads, and cannot be recovered by import. Parameter removal does not guarantee that the service erases an existing credential; set an explicit supported value or replace the guardrail. Check identities and flags remain visible in plans.
