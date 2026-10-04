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
