# Gateway configuration

This complete example manages an organization integration, its workspace access, a provider, and a routing config. Choose an existing workspace and provider family UUID. Upstream credentials are sensitive and are kept outside the visible routing document. Applying it does not send inference traffic.

```hcl
terraform {
  required_providers {
    prisma-airs = { source = "cdot65/prisma-airs", version = "~> 0.9.0" }
  }
}
provider "prisma-airs" {}
variable "workspace_id" { type = string }
variable "ai_provider_id" { type = string }
variable "model" { type = string }
variable "upstream_api_key" {
  type = string
  sensitive = true
}
resource "prisma-airs_gateway_integration" "application" {
  name = "Example - Gateway - Development"
  ai_provider_id = var.ai_provider_id
  key = var.upstream_api_key
}
resource "prisma-airs_gateway_integration_workspace_binding" "application" {
  integration_id = prisma-airs_gateway_integration.application.id
  workspace_id = var.workspace_id
}
resource "prisma-airs_gateway_provider" "application" {
  name = "Example - Provider - Development"
  integration_id = prisma-airs_gateway_integration.application.id
  workspace_id = var.workspace_id
  depends_on = [prisma-airs_gateway_integration_workspace_binding.application]
}
resource "prisma-airs_gateway_config" "application" {
  name = "Example - Routing - Development"
  workspace_id = var.workspace_id
  config = {
    targets = [{
      provider = "@${prisma-airs_gateway_provider.application.slug}"
      override_params = { model = var.model }
    }]
    retry = { attempts = 1 }
  }
}
data "prisma-airs_gateway_configs" "workspace" {
  workspace_id = var.workspace_id
}
output "config_id" { value = prisma-airs_gateway_config.application.id }
output "config_version_id" { value = prisma-airs_gateway_config.application.version_id }
```

Run `terraform init`, `terraform plan`, then `terraform apply`. Change `retry.attempts` to review a config leaf update and computed version change. A second plan after apply should be empty. Destroy removes the owned config/provider/integration and disables its owned workspace binding; it leaves the workspace intact. Protect Terraform state, which contains the upstream secret.

See the [workflow](../guides/gateway-workflow.md) and [exact catalog](../reference/index.md).
