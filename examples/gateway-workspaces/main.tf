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
