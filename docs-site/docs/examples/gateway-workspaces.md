# Gateway workspace and routing

Create a dedicated IAM scope, a Gateway workspace, and a child routing config with provider v0.10.0. This focused project demonstrates the ownership graph. It configures routing; it does not issue an application key or send inference traffic.

## Before you start

Load [shared management credentials](../getting-started/authentication.md) and use an account with Gateway admin/IAM permissions and existing role grants. Scope binding associates the workspace slug; Terraform does not create access policies or workspace membership.

Choose unused names and a dedicated scope. Do not share that managed scope with other resources or states. The literal upstream provider in this minimal configuration illustrates document shape; a real application needs a usable upstream connection. The [expanded Gateway project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-gateway) supplies the connection, guardrails, application keys, and explicit request helpers.

## Configure and apply

Save this complete configuration as `main.tf`:

```hcl
# Setup: Install the workspace-enabled provider.
terraform {
  required_version = ">= 1.11.0, < 2.0.0"
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

# Authentication: Load shared OAuth credentials from the environment.
provider "prisma-airs" {}

# Inputs: Supply tenant-approved workspace metadata.
variable "workspace_metadata" {
  description = "Tenant-approved metadata including every required workspace property."
  type        = map(string)
}

# Workspace: Own a dedicated scope and complete workspace settings.
resource "prisma-airs_gateway_workspace" "applications" {
  name             = "Application gateway"
  scope_name       = "tf_gateway_apps_prod"
  scope_management = "managed"

  defaults = {
    metadata = var.workspace_metadata
  }
  rate_limits = [
    {
      type  = "requests"
      unit  = "rpm"
      value = 60
    },
    {
      type  = "tokens"
      unit  = "tpm"
      value = 1000
    },
  ]
  usage_limits = [
    {
      type         = "tokens"
      credit_limit = 100000
    }
  ]
}

# Routing: Reference the workspace UUID so children are destroyed first.
resource "prisma-airs_gateway_config" "applications" {
  name         = "Application routing"
  workspace_id = prisma-airs_gateway_workspace.applications.id
  config = {
    provider = "openai"
    retry = {
      attempts = 1
    }
  }
}

# Discovery: Read safe metadata without inferring inventory completeness.
data "prisma-airs_gateway_workspace" "applications" {
  workspace_id = prisma-airs_gateway_workspace.applications.id
}

# Outputs: Pass identifiers to your application configuration workflow.
output "workspace_uuid" {
  value = prisma-airs_gateway_workspace.applications.id
}
output "workspace_slug" {
  value = data.prisma-airs_gateway_workspace.applications.slug
}
```

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output
```

Supply `TF_VAR_workspace_metadata` as a map of tenant-approved keys and values, including every required workspace property. Use `{}` only if the tenant permits empty metadata. Missing or arbitrary properties can produce HTTP 400; there is no universal valid metadata object. The workspace owns its configured metadata and policy collections. The service permits one usage policy per workspace. Do not also create a standalone usage policy for this same workspace. Do not reference a child config in workspace defaults during creation; that would form a Terraform dependency cycle. Set an existing config default in a later apply.

## Update and clean up

Rename `name` to change the display label while retaining the UUID and slug. Change the inline credit or rate settings to inspect native plan diffs. Removing configured collections/defaults clears those settings; never configured settings remain unmanaged. Scope mode/name changes replace the workspace.

```bash
terraform plan -out=update.tfplan
terraform apply update.tfplan
terraform plan -detailed-exitcode
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Destroy deletes the child config, archives the workspace, and confirms deletion of its dedicated scope. Historical workspace rows can remain. Scope deletion is independent of workspace archival.

For an externally owned scope, set `scope_management = "external"` and its existing name **before the first apply**. External owners maintain bindings and role grants; Terraform makes no IAM writes and archives only its workspace on destroy. For import/adoption, partial failure checkpoints, and recovery, read the [workspace lifecycle guide](../resources/gateway-workspace.md). Matching names or incomplete inventory are not sufficient ownership evidence.
