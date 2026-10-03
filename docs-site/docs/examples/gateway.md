# Gateway configuration

Create an owned upstream integration, authorize an existing workspace to use it, and save a single-model routing configuration. This focused example teaches the dependency chain; Terraform apply does not send inference traffic.

## Before you start

Install [provider v0.9.0](../getting-started/installation.md) and load the three [management environment variables](../getting-started/authentication.md). You also need an existing Gateway workspace UUID, a provider-family UUID, and a model available through that connection.

Supply `workspace_id`, `ai_provider_id`, and `model` through a nonsecret `terraform.tfvars` file. Load the real upstream key through `TF_VAR_upstream_api_key` from your secret store. The model service and workspace are external prerequisites; model enablement is outside this provider's scope.

## Configure the connection and routing

Save this complete configuration as `main.tf` in a separate directory. Choose unused names for the owned objects.

```hcl
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
```

The integration holds the sensitive upstream key. The workspace binding grants access, and the provider waits for that binding. The routing document selects that provider directly, sets the model, and bounds retries. A multi-target routing document also needs its routing strategy.

## Apply and inspect

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output
```

`config_id` identifies the routing configuration; `config_version_id` identifies its current document version. Protect state and saved plans, which can contain the upstream key. A successful apply establishes configuration management, not inference connectivity.

## Make a change and clean up

Change `retry.attempts`, review a saved plan, and apply it. The config ID stays stable while its version changes. A subsequent unchanged plan should exit 0:

```bash
terraform plan -out=update.tfplan
terraform apply update.tfplan
terraform plan -detailed-exitcode
```

Remove the owned configuration using the same inputs, tenant, and state:

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Destroy removes the config, provider, and integration and disables the owned workspace binding. The existing workspace remains external.

## Build a governed application

For AIRS inspection, application keys, four routing lessons, request/token policies, MCP, and other optional platform features, follow the [expanded Gateway project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-gateway). Its getting-started guide includes explicit request helpers and recorded live results.

See the [Gateway workflow](../guides/gateway-workflow.md) for lifecycle details and the [exact catalog](../reference/index.md) for supported attributes.
