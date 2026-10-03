# Setup: Declare the provider required by this configuration.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.9.0"
    }
  }
}

# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {}

# Inputs: Supply environment-specific values; load sensitive values from your secret store.
variable "workspace_id" {
  type = string
}

variable "ai_provider_id" {
  type = string
}

variable "model" {
  type = string
}

variable "upstream_api_key" {
  type      = string
  sensitive = true
}

# Connection: Keep upstream credentials in the integration, outside routing.
resource "prisma-airs_gateway_integration" "application" {
  name           = "Example - Gateway - Development"
  ai_provider_id = var.ai_provider_id
  key            = var.upstream_api_key
}

# Workspace access: Authorize this workspace to use the owned integration.
resource "prisma-airs_gateway_integration_workspace_binding" "application" {
  integration_id = prisma-airs_gateway_integration.application.id
  workspace_id   = var.workspace_id
}

# Provider: Expose the integration after its workspace binding exists.
resource "prisma-airs_gateway_provider" "application" {
  name           = "Example - Provider - Development"
  integration_id = prisma-airs_gateway_integration.application.id
  workspace_id   = var.workspace_id
  depends_on     = [prisma-airs_gateway_integration_workspace_binding.application]
}

# Routing: Keep the visible routing document free of upstream credentials.
resource "prisma-airs_gateway_config" "application" {
  name         = "Example - Routing - Development"
  workspace_id = var.workspace_id

  config = {
    provider = "@${prisma-airs_gateway_provider.application.slug}"

    override_params = {
      model = var.model
    }

    retry = {
      attempts = 1
    }
  }
}

# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_gateway_configs" "workspace" {
  workspace_id = var.workspace_id
}

# Outputs: Expose results for the next configuration or application step.
output "config_id" {
  value = prisma-airs_gateway_config.application.id
}

output "config_version_id" {
  value = prisma-airs_gateway_config.application.version_id
}
