# Gateway workspace and routing

Use Terraform 1.11+ and the pinned provider 0.10.0. Load shared management credentials for an entitled tenant with Gateway admin/IAM role grants. Choose an unused name and dedicated `scope_name` before applying.

Supply `TF_VAR_workspace_metadata` as a map of tenant-approved keys/values including all required properties. Empty metadata works only if permitted by your tenant; arbitrary keys or missing required values can return HTTP 400.

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output
```

This configuration owns its workspace and child routing config. Inline policies belong to the workspace; do not also create a standalone usage policy for it. The literal upstream family illustrates the routing document; it does not provision a model connection or application key. Use the [complete application project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-gateway) for inference.

Change the display name or inline policy settings, review a saved plan, and apply. Destroy removes the child, archives the workspace, then deletes its owned scope. Terraform creates no access policies or membership. See the [workspace lesson](https://cdot65.github.io/terraform-provider-prisma-airs/examples/gateway-workspaces/) for external-scope ownership, import, and recovery.
