---
page_title: "prisma-airs_gateway_workspace (Data Source)"
subcategory: "AI Gateway"
---

# prisma-airs_gateway_workspace Data Source

`prisma-airs_gateway_workspace` reads safe workspace metadata through the tenant admin plane.

```hcl
# Discovery: Read metadata for an existing workspace UUID.
data "prisma-airs_gateway_workspace" "applications" {
  workspace_id = var.workspace_id
}
```

Returns UUID, label, slug, scope name when supplied, lifecycle status, timestamps and default-workspace indicator. Defaults, security settings, credentials and users are excluded. A missing or archived workspace returns an error; this lookup never creates or binds an IAM scope.

See the [exact lookup schema](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/data-source-prisma-airs_gateway_workspace/) and [managed workspace guide](https://cdot65.github.io/terraform-provider-prisma-airs/resources/gateway-workspace/).

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Remote workspace created_at when available. |
| `description` | `string` | computed | — | Remote workspace description when available. |
| `id` | `string` | computed | — | Remote workspace id when available. |
| `is_default` | `bool` | computed | — | Whether this is the tenant default workspace. |
| `last_updated_at` | `string` | computed | — | Remote workspace last_updated_at when available. |
| `name` | `string` | computed | — | Remote workspace name when available. |
| `scope_name` | `string` | computed | — | Remote workspace scope_name when available. |
| `slug` | `string` | computed | — | Remote workspace slug when available. |
| `status` | `string` | computed | — | Remote workspace status when available. |
| `workspace_id` | `string` | required | — | Workspace UUID to read. |
