# Gateway workspace and routing

This complete configuration creates a dedicated IAM scope, a Gateway workspace and a child routing config. Load the shared management credentials and use an account with Gateway admin/IAM permissions and existing role grants. Terraform does not create access policies or workspace membership.

```hcl
# Setup: Install the workspace-enabled provider.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

provider "prisma-airs" {}

resource "prisma-airs_gateway_workspace" "applications" {
  name             = "Application gateway"
  scope_name       = "tf_gateway_apps_prod"
  scope_management = "managed"

  defaults = { metadata = { owner = "terraform" } }
  rate_limits = [
    { type = "requests", unit = "rpm", value = 60 },
    { type = "tokens", unit = "tpm", value = 1000 },
  ]
  usage_limits = [{ type = "tokens", credit_limit = 100000 }]
}

resource "prisma-airs_gateway_config" "applications" {
  name         = "Application routing"
  workspace_id = prisma-airs_gateway_workspace.applications.id
  config       = { provider = "openai", retry = { attempts = 1 } }
}

data "prisma-airs_gateway_workspace" "applications" {
  workspace_id = prisma-airs_gateway_workspace.applications.id
}

output "workspace_uuid" {
  value = prisma-airs_gateway_workspace.applications.id
}
output "workspace_slug" {
  value = data.prisma-airs_gateway_workspace.applications.slug
}
```

Run `terraform init`, `terraform plan`, then `terraform apply`. Rename `name` to update the label while retaining UUID/slug. Remove the configured policy collections/defaults to clear them. On destroy, the child config is deleted first, the workspace is archived, and the dedicated IAM scope is deleted with independent confirmation.

For an existing externally owned scope, change `scope_management` to `external` before the first apply and supply its name. External owners maintain bindings and role grants; Terraform makes no IAM writes. Import/adoption and partial-failure instructions are in the [workspace guide](../resources/gateway-workspace.md).
