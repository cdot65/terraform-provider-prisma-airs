# Native targets

This complete configuration illustrates all seven connection families. Substitute endpoints and model identifiers you own and supply sensitive variables through your secret store. Every target has exactly one connection block.

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source = "cdot65/prisma-airs"
    }
  }
}

provider "prisma-airs" {}

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

resource "prisma-airs_red_team_target" "openai" {
  name = "terraform-openai"
  openai {
    api_key       = var.model_api_key
    model_name    = "your-openai-model"
    request_body  = { messages = [{ role = "user", content = "{INPUT}" }] }
    response_body = { choices = [{ message = { content = "{RESPONSE}" } }] }
    response_key  = "content"
  }
}

resource "prisma-airs_red_team_target" "hugging_face" {
  name = "terraform-hugging-face"
  hugging_face {
    api_key       = var.model_api_key
    model_name    = "your-organization/your-model"
    request_body  = { inputs = "{INPUT}" }
    response_body = { generated_text = "{RESPONSE}" }
    response_key  = "generated_text"
  }
}

resource "prisma-airs_red_team_target" "databricks" {
  name = "terraform-databricks"
  databricks {
    workspace_url       = "https://workspace.cloud.databricks.com"
    model_name          = "your-serving-endpoint"
    access_token        = var.databricks_token
    request_body        = { messages = [{ role = "user", content = "{INPUT}" }], stream = true }
    response_body       = { choices = [{ delta = { content = "{RESPONSE}" } }] }
    response_key        = "content"
    response_stop_key   = "finish_reason"
    response_stop_value = "stop"
  }
}

resource "prisma-airs_red_team_target" "bedrock" {
  name = "terraform-bedrock"
  bedrock {
    access_id     = var.aws_access_id
    access_secret = var.aws_access_secret
    region        = "us-west-2"
    model_id      = "your-bedrock-model-id"
    request_body  = { messages = [{ role = "user", content = [{ text = "{INPUT}" }] }] }
    response_body = { output = { message = { content = [{ text = "{RESPONSE}" }] } } }
    response_key  = "text"
  }
}

resource "prisma-airs_red_team_target" "custom" {
  name = "terraform-custom"
  custom {
    api_endpoint  = "https://app.example.com/chat"
    request_body  = { prompt = "{INPUT}" }
    response_body = { answer = "{RESPONSE}" }
    response_key  = "answer"
  }
}

resource "prisma-airs_red_team_target" "rest" {
  name = "terraform-rest"
  rest {
    api_endpoint    = "https://app.example.com/chat"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { prompt = "{INPUT}", temperature = 0, stream = false, context = null }
    response_body   = { answer = "{RESPONSE}" }
    response_key    = "answer"
  }
  headers_auth {
    headers = { Authorization = "Bearer ${var.app_token}" }
  }
}

resource "prisma-airs_red_team_target" "streaming" {
  name = "terraform-streaming"
  streaming {
    api_endpoint        = "https://app.example.com/stream"
    request_body        = { prompt = "{INPUT}", stream = true }
    response_body       = { token = "{RESPONSE}" }
    response_key        = "token"
    response_stop_key   = "event"
    response_stop_value = "done"
  }
}
```

Create and update register management configuration without inference or connection validation. Native provider setting changes plan replacement; custom transport metadata and supported mode/auth-method changes can update in place. Read the [target lifecycle contract](../resources/red-team-target.md).

Databricks uses streaming responses and requires stop key/value settings. Choose either `access_token` or OAuth `client_id` and `secret`. The latter are distinct from this provider's management OAuth credentials.

To attach a Network Broker, set `api_endpoint_type = "NETWORK_BROKER"` and a preexisting `network_broker_channel_uuid`. The channel stays externally managed.
