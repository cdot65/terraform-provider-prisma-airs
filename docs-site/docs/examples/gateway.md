# OpenAI GPT and Anthropic Claude Opus

Create two upstream connections in an existing Gateway workspace without supplying provider-family UUIDs. Terraform discovers the catalog, selects `open-ai` and `anthropic`, then creates workspace bindings, providers, model routes, and application keys.

## Before you start

This example requires provider **0.11.0**, which includes `prisma-airs_gateway_ai_providers`. Install it from the [Terraform Registry](../getting-started/installation.md#install-from-the-terraform-registry); no local provider build is required.

Load the three [management environment variables](../getting-started/authentication.md). You also need an existing Gateway workspace, a Gateway inference deployment, and OpenAI and Anthropic **API** credentials with access to the selected models. A ChatGPT subscription is separate from OpenAI API access; this example routes GPT chat completions through the OpenAI API.

The sample model IDs are `gpt-4.1` and `claude-opus-4-6`. They are explicit API IDs, not provider slugs or a promise of the latest model. Check [OpenAI models](https://developers.openai.com/api/docs/models/gpt-4.1) and [Claude model IDs](https://platform.claude.com/docs/en/about-claude/models/model-ids-and-versions), then select models available to your account and enabled in Gateway. The provider catalog does not enable models or verify upstream billing/access.

## Preview real provider discovery

Before supplying model-service credentials, run the [read-only catalog configuration](../data-sources/gateway-ai-providers.md#run-discovery-before-configuring-models). It uses only management authentication. This is actual Terraform output from the development-provider run on 2026-10-04, with UUID values sanitized:

```text
catalog_counts = {
  "active" = 79
  "returned" = 81
}
provider_family_ids = {
  "anthropic" = "<anthropic-provider-family-uuid>"
  "open-ai" = "<openai-provider-family-uuid>"
}
```

Terraform reported `Resources: 0 added, 0 changed, 0 destroyed.` The [full CLI transcript](https://github.com/cdot65/prisma-airs-terraform-examples/blob/main/docs/live-runs/gateway-provider-catalog.txt) records the unchanged plan and cleanup. These values came from the live catalog, not mocked tests; counts and UUIDs can differ in your environment.

## Configure both connections

Save the following complete configuration as `main.tf` in a new directory. Set `workspace_id` and an unused `name_prefix` in a nonsecret `terraform.tfvars` file. Override `models` only when your environment uses different model IDs. There is no provider UUID input.

Load `TF_VAR_upstream_api_keys` from your credential store as a JSON map with this shape; keep actual keys out of input files and terminal output:

```json
{"open-ai": "<openai-api-key>", "anthropic": "<anthropic-api-key>"}
```

```hcl
# Setup: Pin the release that includes provider-family discovery.
terraform {
  required_version = ">= 1.11.0, < 2.0.0"

  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "= 0.11.0"
    }
  }
}

# Authentication: Load management credentials through PANW_MGMT_*.
provider "prisma-airs" {}

variable "workspace_id" {
  description = "Existing Gateway workspace UUID; this project does not own it."
  type        = string
}

variable "name_prefix" {
  description = "Unused prefix for this project's owned connections and routes."
  type        = string
  default     = "tf-catalog-chat"
}

variable "upstream_api_keys" {
  description = "Upstream API credentials keyed by open-ai and anthropic; load from the environment."
  type        = map(string)
  sensitive   = true

  validation {
    condition = alltrue([
      for slug in ["open-ai", "anthropic"] :
      try(length(trimspace(var.upstream_api_keys[slug])) > 0, false)
    ])
    error_message = "Supply both open-ai and anthropic API credentials through TF_VAR_upstream_api_keys."
  }
}

variable "models" {
  description = "API model IDs enabled in Gateway and available to each upstream account."
  type        = map(string)
  default = {
    open-ai   = "gpt-4.1"
    anthropic = "claude-opus-4-6"
  }

  validation {
    condition = toset(keys(var.models)) == toset(["open-ai", "anthropic"]) && alltrue([
      for model in values(var.models) : length(trimspace(model)) > 0
    ])
    error_message = "Set exactly open-ai and anthropic, each with a nonempty API model ID."
  }
}

# Discovery: Resolve active provider-family UUIDs by exact catalog slug.
data "prisma-airs_gateway_ai_providers" "catalog" {}

# Connections: Terraform owns these integrations, not the catalog entries.
resource "prisma-airs_gateway_integration" "chat" {
  for_each       = var.models
  name           = "${var.name_prefix}-${each.key}"
  ai_provider_id = data.prisma-airs_gateway_ai_providers.catalog.ids_by_slug[each.key]
  key            = var.upstream_api_keys[each.key]
}

# Workspace access: Authorize the existing workspace before creating providers.
resource "prisma-airs_gateway_integration_workspace_binding" "chat" {
  for_each       = var.models
  integration_id = prisma-airs_gateway_integration.chat[each.key].id
  workspace_id   = var.workspace_id
}

resource "prisma-airs_gateway_provider" "chat" {
  for_each       = var.models
  name           = "${var.name_prefix}-${each.key}"
  integration_id = prisma-airs_gateway_integration.chat[each.key].id
  workspace_id   = var.workspace_id
  depends_on     = [prisma-airs_gateway_integration_workspace_binding.chat]
}

# Routing: Model IDs select GPT or Claude within their respective connections.
resource "prisma-airs_gateway_config" "chat" {
  for_each     = var.models
  name         = "${var.name_prefix}-${each.key}"
  workspace_id = var.workspace_id

  config = {
    provider = "@${prisma-airs_gateway_provider.chat[each.key].slug}"

    override_params = {
      model = each.value
    }

    retry = {
      attempts = 1
    }
  }
}

# Application access: Give each model route its own scoped credential.
resource "prisma-airs_gateway_service_api_key" "chat" {
  for_each     = var.models
  name         = "${var.name_prefix}-${each.key}"
  workspace_id = var.workspace_id
  scopes       = ["completions.write"]

  defaults = {
    config_id             = prisma-airs_gateway_config.chat[each.key].id
    allow_config_override = false
  }
}

# Outputs: Return UUIDs and selected models without exposing upstream keys.
output "routes" {
  value = {
    for slug, model in var.models : slug => {
      ai_provider_id = data.prisma-airs_gateway_ai_providers.catalog.ids_by_slug[slug]
      integration_id = prisma-airs_gateway_integration.chat[slug].id
      provider_id    = prisma-airs_gateway_provider.chat[slug].id
      config_id      = prisma-airs_gateway_config.chat[slug].id
      model          = model
    }
  }
}

output "application_keys" {
  description = "Sensitive Gateway keys; protect state and saved plans."
  sensitive   = true
  value       = { for slug, key in prisma-airs_gateway_service_api_key.chat : slug => key.key }
}
```

The lookup uses exact catalog slugs, not display names: OpenAI is `open-ai`; Anthropic is `anthropic`. `ids_by_slug` contains active services only. Missing or inactive entries fail lookup, and duplicate identities fail discovery instead of silently choosing a connection. The catalog data source stores metadata in state but owns no remote object.

The integration owns the upstream connection, the binding authorizes the existing workspace, and the workspace provider exposes that connection to routing. Each routing document selects a model and has its own scoped application key. Terraform owns ten resources in total; the existing workspace and catalog remain external.

## Apply and inspect

With management credentials and model inputs configured:

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output routes
```

`routes` distinguishes the provider-family UUID, organization integration UUID, workspace provider UUID, routing config UUID, and selected model. The sensitive `application_keys` output contains Gateway credentials, not upstream API keys. Protect state and saved plans.

A successful apply establishes configuration; it does not prove that a model is enabled or callable.

## Call GPT and Claude

Use the [public catalog example](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-gateway/provider-catalog) for copyable request commands for each route, expected outcomes, and cleanup. The inference endpoint and Gateway application keys are separate from management credentials. Enable/register the selected models in Gateway if needed before sending traffic.

## A real inference response

A supplemental run discovered existing connections with `airs cli aigateway`, then used Terraform to create temporary configs and service application keys. A standard OpenAI integration returned this actual response excerpt:

```json
{
  "model": "gpt-4.1-2025-04-14",
  "choices": [{"message": {"content": "Hello!", "role": "assistant"}}],
  "usage": {"completion_tokens": 2, "prompt_tokens": 13, "total_tokens": 15}
}
```

See the [full recorded response and exact Terraform source](https://github.com/cdot65/prisma-airs-terraform-examples/blob/main/docs/live-runs/gateway-existing-models.md). This excerpt is real output, not a mock or a full response schema. The run reused an existing connection; it did not apply the ten-resource project above that creates new upstream integrations.

Claude Opus requests through all three existing Vertex connections returned HTTP 401 authentication errors. The existing Bedrock connection returned HTTP 403 with an invalid security token. No direct Anthropic integration was configured, so successful Claude inference remains unverified. The recorded evidence includes the actual errors and cleanup; supply valid upstream credentials before expecting the direct Anthropic route to work. All temporary configs and keys were destroyed.

## Change and clean up

Change a value in `models`, review `terraform plan`, and apply the routing update. Config IDs stay stable while their revisions change. After model enablement, send another request to verify the selected model.

Destroy using the same inputs, tenant, and state:

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Destroy removes both model routes, application keys, workspace providers, and integrations, and disables their owned workspace bindings. It does not delete the existing workspace or catalog entries.

For existing shared connections, read [Gateway integrations](../data-sources/gateway-integrations.md) instead of recreating them. For guardrails, request budgets, and four routing lessons, see the [expanded Gateway project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-gateway), which retains its provider 0.10.0 compatibility configuration.
