# AI Gateway workspace lookup

`prisma-airs_gateway_workspace` reads safe workspace metadata through the tenant admin plane.

```hcl
# Discovery: Read metadata for an existing workspace UUID.
data "prisma-airs_gateway_workspace" "applications" {
  workspace_id = var.workspace_id
}
```

Returns UUID, label, slug, scope name when supplied, lifecycle status, timestamps and default-workspace indicator. Defaults, security settings, credentials and users are excluded. A missing or archived workspace returns an error; this lookup never creates or binds an IAM scope.

See the [exact lookup schema](../reference/generated/data-source-prisma-airs_gateway_workspace.md) and [managed workspace guide](../resources/gateway-workspace.md).
