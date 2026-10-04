# Native targets

This reference configuration illustrates all seven connection families. Each target has exactly one connection block, with native request and response objects that retain meaningful false, zero, and null values.

## Before you start

Install the [released provider](../getting-started/installation.md), load management credentials, and prepare endpoints and model identifiers you control. Load sensitive `TF_VAR_*` strings through your secret store; model-service credentials differ from management OAuth.

For a first exercise, keep one target and its required variable declarations **before the first apply**. Applying the entire reference requires credentials and access for every service shown. Removing targets from an existing managed configuration proposes their destruction, so review the plan.

## Choose a connection family

```hcl
# Setup: Declare the provider required by this configuration.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {}

# Inputs: Supply environment-specific values; load sensitive values from your secret store.
variable "model_api_key" {
  type      = string
  sensitive = true
}

variable "databricks_token" {
  type      = string
  sensitive = true
}

variable "aws_access_id" {
  type      = string
  sensitive = true
}

variable "aws_access_secret" {
  type      = string
  sensitive = true
}

variable "app_token" {
  type      = string
  sensitive = true
}

# OpenAI: Supply model credentials and match the chat payload.
resource "prisma-airs_red_team_target" "openai" {
  name = "terraform-openai"

  openai {
    api_key    = var.model_api_key
    model_name = "your-openai-model"

    request_body = {
      messages = [
        {
          role    = "user"
          content = "{INPUT}"
        }
      ]
    }

    response_body = {
      choices = [
        {
          message = {
            content = "{RESPONSE}"
          }
        }
      ]
    }

    response_key = "content"
  }
}

# Hugging Face: Match the model input and output templates.
resource "prisma-airs_red_team_target" "hugging_face" {
  name = "terraform-hugging-face"

  hugging_face {
    api_key    = var.model_api_key
    model_name = "your-organization/your-model"

    request_body = {
      inputs = "{INPUT}"
    }

    response_body = {
      generated_text = "{RESPONSE}"
    }

    response_key = "generated_text"
  }
}

# Databricks: Streaming responses require stop-key and stop-value settings.
resource "prisma-airs_red_team_target" "databricks" {
  name = "terraform-databricks"

  databricks {
    workspace_url = "https://workspace.cloud.databricks.com"
    model_name    = "your-serving-endpoint"
    access_token  = var.databricks_token

    request_body = {
      messages = [
        {
          role    = "user"
          content = "{INPUT}"
        }
      ]

      stream = true
    }

    response_body = {
      choices = [
        {
          delta = {
            content = "{RESPONSE}"
          }
        }
      ]
    }

    response_key        = "content"
    response_stop_key   = "finish_reason"
    response_stop_value = "stop"
  }
}

# Bedrock: Use model-service credentials distinct from management OAuth.
resource "prisma-airs_red_team_target" "bedrock" {
  name = "terraform-bedrock"

  bedrock {
    access_id     = var.aws_access_id
    access_secret = var.aws_access_secret
    region        = "us-west-2"
    model_id      = "your-bedrock-model-id"

    request_body = {
      messages = [
        {
          role = "user"

          content = [
            {
              text = "{INPUT}"
            }
          ]
        }
      ]
    }

    response_body = {
      output = {
        message = {
          content = [
            {
              text = "{RESPONSE}"
            }
          ]
        }
      }
    }

    response_key = "text"
  }
}

# Custom endpoint: Describe the request and response payloads you control.
resource "prisma-airs_red_team_target" "custom" {
  name = "terraform-custom"

  custom {
    api_endpoint = "https://app.example.com/chat"

    request_body = {
      prompt = "{INPUT}"
    }

    response_body = {
      answer = "{RESPONSE}"
    }

    response_key = "answer"
  }
}

# REST: Preserve false, zero, and null values in native payload objects.
resource "prisma-airs_red_team_target" "rest" {
  name = "terraform-rest"

  rest {
    api_endpoint = "https://app.example.com/chat"

    request_headers = {
      "Content-Type" = "application/json"
    }

    request_body = {
      prompt      = "{INPUT}"
      temperature = 0
      stream      = false
      context     = null
    }

    response_body = {
      answer = "{RESPONSE}"
    }

    response_key = "answer"
  }

  headers_auth {
    headers = {
      Authorization = "Bearer ${var.app_token}"
    }
  }
}

# Streaming: Describe incremental output and the event that ends the response.
resource "prisma-airs_red_team_target" "streaming" {
  name = "terraform-streaming"

  streaming {
    api_endpoint = "https://app.example.com/stream"

    request_body = {
      prompt = "{INPUT}"
      stream = true
    }

    response_body = {
      token = "{RESPONSE}"
    }

    response_key        = "token"
    response_stop_key   = "event"
    response_stop_value = "done"
  }
}
```

## Apply and use the target

Save the chosen configuration as `main.tf`, then review and apply:

```bash
terraform init
terraform validate
terraform plan -out=targets.tfplan
terraform apply targets.tfplan
```

Create and update register management configuration without inference or connection validation. Native provider setting changes plan replacement; custom transport metadata and supported mode/auth-method changes can update in place. Read the [target lifecycle contract](../resources/red-team-target.md).

Databricks uses streaming responses and requires stop key/value settings. Choose either `access_token` or OAuth `client_id` and `secret`. The latter are distinct from this provider's management OAuth credentials.

To attach a Network Broker, set `api_endpoint_type = "NETWORK_BROKER"` and a preexisting `network_broker_channel_uuid`. The channel stays externally managed.

## Clean up

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Only the owned targets are removed; the model services and application endpoints remain external. Protect state because it contains connection credentials. For a focused application target and prompt-set project, use the [Red Teaming project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-red-teaming).
