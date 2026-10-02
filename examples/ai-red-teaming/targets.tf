# Native HCL target configuration; no inference runs during apply.
# Network Broker requires an explicit preexisting channel; these examples use PUBLIC.
# The legacy litellm_multiturn address is preserved; multi-turn settings are outside this schema.

resource "prisma-airs_red_team_target" "litellm_multiturn" {
  name = "LiteLLM configuration"
  rest {
    api_endpoint    = "https://litellm.cdot.io/v1/chat/completions"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { model = "mistral-7b", messages = [{ role = "user", content = "{INPUT}" }], max_tokens = 256 }
    response_body   = { choices = [{ message = { content = "{RESPONSE}" } }] }
    response_key    = "content"
  }
  headers_auth {
    headers = { "Authorization" = "Bearer ${var.litellm_api_key}" }
  }
}

resource "prisma-airs_red_team_target" "litellm_singleturn" {
  name = "LiteLLM single-turn"
  rest {
    api_endpoint    = "https://litellm.cdot.io/v1/chat/completions"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { model = "mistral-7b", messages = [{ role = "user", content = "{INPUT}" }], max_tokens = 256 }
    response_body   = { choices = [{ message = { content = "{RESPONSE}" } }] }
    response_key    = "content"
  }
  headers_auth {
    headers = { "Authorization" = "Bearer ${var.litellm_api_key}" }
  }
}

resource "prisma-airs_red_team_target" "worf" {
  name = "worf - local"
  rest {
    api_endpoint    = "https://redteam.cdot.io/api/v1/litellm/chat/completions"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { model = "qwen3-14b-awq", messages = [{ role = "user", content = "{INPUT}" }], max_tokens = 256 }
    response_body   = { choices = [{ message = { content = "{RESPONSE}" } }] }
    response_key    = "content"
  }
  headers_auth {
    headers = { "X-Api-Key" = var.worf_api_key }
  }
}

resource "prisma-airs_red_team_target" "qwen_completions" {
  name = "Qwen completions"
  rest {
    api_endpoint    = "https://uoft-qwen2-5-7b-instruct-8828s.paas.ai.telus.com/v1/completions"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { model = "Qwen/Qwen2.5-7B-Instruct", prompt = "{INPUT}", max_tokens = 300 }
    response_body   = { choices = [{ text = "{RESPONSE}" }] }
    response_key    = "text"
  }
  headers_auth {
    headers = { "Authorization" = "Bearer ${var.qwen_api_key}" }
  }
}

resource "prisma-airs_red_team_target" "talkdesk" {
  name = "Talkdesk"
  rest {
    api_endpoint    = "https://redteam.cdot.io/api/v1/talkdesk/prompt"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { prompt = "{INPUT}", contact_name = "Test", subject = "Account help" }
    response_body   = { response = "{RESPONSE}" }
    response_key    = "response"
  }
  headers_auth {
    headers = { "X-Api-Key" = var.talkdesk_api_key }
  }
}

resource "prisma-airs_red_team_target" "bedrock_claude" {
  name = "AWS Bedrock - Claude"
  bedrock {
    access_id     = var.bedrock_access_id
    access_secret = var.bedrock_access_secret
    region        = "us-west-2"
    model_id      = "us.anthropic.claude-opus-4-6-v1"
    api_endpoint  = "https://bedrock-runtime.us-west-2.amazonaws.com/model/us.anthropic.claude-opus-4-6-v1/converse"
    request_body = {
      messages        = [{ role = "user", content = [{ text = "{INPUT}" }] }]
      inferenceConfig = { maxTokens = 512, temperature = 0.7 }
    }
    response_body = { output = { message = { content = [{ text = "{RESPONSE}" }] } } }
    response_key  = "text"
  }
}
