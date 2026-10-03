# Gateway workspace and routing

Load shared management credentials for an entitled tenant with existing Gateway admin/IAM role grants. Run `terraform init`, `terraform plan`, and `terraform apply`. Use a unique dedicated `scope_name`.

This example creates its workspace and child config. Destroy removes the child, archives the workspace, then deletes the owned scope. No access policies or membership are managed. See the [workspace guide](https://cdot65.github.io/terraform-provider-prisma-airs/resources/gateway-workspace/) for import and recovery.
