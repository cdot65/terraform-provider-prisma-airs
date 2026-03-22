# ---------------------------------------------------------------------------
# Red Team Targets
# ---------------------------------------------------------------------------
# Real-world targets spanning different connection types, model providers,
# and deployment patterns. Each target demonstrates a different integration.
# ---------------------------------------------------------------------------

# ── LiteLLM (Mistral-7b) — Network Broker, Multi-turn ────────────────
# Self-hosted LiteLLM proxy serving Mistral-7b via OpenAI-compatible API.
# Connected through AIRS network broker. Multi-turn enabled.

resource "prisma-airs_red_team_target" "litellm_multiturn" {
  name            = "litellm.cdot.io - no guardrails - REST APIv2"
  target_type     = "APPLICATION"
  connection_type = "CUSTOM"

  connection_params = jsonencode({
    api_endpoint = "https://litellm.cdot.io/v1/chat/completions"
    request_headers = {
      "Content-Type"  = "application/json"
      "Authorization" = "Bearer ${var.litellm_api_key}"
    }
    request_json = {
      model = "mistral-7b"
      messages = [
        {
          role    = "user"
          content = "{INPUT}"
        }
      ]
      max_tokens = 256
    }
    response_json = {
      choices = [
        {
          message = {
            content = "{RESPONSE}"
          }
        }
      ]
    }
    response_key = "content"
    multi_turn_config = {
      type           = "stateless"
      assistant_role = "assistant"
    }
  })
}

# ── LiteLLM (Mistral-7b) — Network Broker, Single-turn ───────────────
# Same LiteLLM proxy but configured for single-turn testing only.
# Useful for baseline prompt injection and jailbreak assessments.

resource "prisma-airs_red_team_target" "litellm_singleturn" {
  name            = "litellm.cdot.io - no guardrails - REST API"
  target_type     = "APPLICATION"
  connection_type = "CUSTOM"

  connection_params = jsonencode({
    api_endpoint = "https://litellm.cdot.io/v1/chat/completions"
    request_headers = {
      "Content-Type"  = "application/json"
      "Authorization" = "Bearer ${var.litellm_api_key}"
    }
    request_json = {
      model = "mistral-7b"
      messages = [
        {
          role    = "user"
          content = "{INPUT}"
        }
      ]
      max_tokens = 256
    }
    response_json = {
      choices = [
        {
          message = {
            content = "{RESPONSE}"
          }
        }
      ]
    }
    response_key = "content"
  })
}

# ── Worf (Qwen3-14B-AWQ) — Local vLLM via Network Broker ─────────────
# Locally-hosted Qwen3-14B-AWQ model served via vLLM behind a LiteLLM
# proxy. General-purpose AI app with broad language and tool support.
# Has AIRS guardrails configured (content filter enabled).

resource "prisma-airs_red_team_target" "worf" {
  name            = "worf - local"
  target_type     = "APPLICATION"
  connection_type = "CUSTOM"

  connection_params = jsonencode({
    api_endpoint = "https://redteam.cdot.io/api/v1/litellm/chat/completions"
    request_headers = {
      "Content-Type" = "application/json"
      "X-Api-Key"    = var.worf_api_key
    }
    request_json = {
      model = "qwen3-14b-awq"
      messages = [
        {
          role    = "user"
          content = "{INPUT}"
        }
      ]
      max_tokens = 2048
    }
    response_json = {
      choices = [
        {
          message = {
            content = "{RESPONSE}"
          }
        }
      ]
    }
    response_key = "content"
    multi_turn_config = {
      type           = "stateless"
      assistant_role = "assistant"
    }
  })
}

# ── AWS Bedrock (Claude Opus 4.6) — Direct Bedrock Streaming ──────────
# Claude Opus 4.6 via AWS Bedrock ConverseStream API. Uses MODEL target
# type (not APPLICATION). Streaming response mode. Configured as a
# recipe chat assistant for focused red team scenarios.

resource "prisma-airs_red_team_target" "bedrock_claude" {
  name            = "AWS Bedrock - Claude 4.6"
  target_type     = "MODEL"
  connection_type = "BEDROCK"

  connection_params = jsonencode({
    api_endpoint = "https://bedrock-runtime.us-west-2.amazonaws.com/model/us.anthropic.claude-opus-4-6-v1/converse-stream"
    request_headers = {
      "Content-Type" = "application/json"
    }
    request_json = {
      messages = [
        {
          role = "user"
          content = [
            { text = "{INPUT}" }
          ]
        }
      ]
      inferenceConfig = {
        maxTokens   = 512
        temperature = 0.7
      }
    }
    response_json = {
      output = {
        message = {
          content = [
            { text = "{RESPONSE}" }
          ]
        }
      }
    }
    response_key = "text"
    target_connection_config = {
      access_id     = var.bedrock_access_id
      access_secret = var.bedrock_access_secret
      region        = "us-west-2"
      model_id      = "us.anthropic.claude-opus-4-6-v1"
    }
    multi_turn_config = {
      type           = "stateless"
      assistant_role = "assistant"
    }
  })
}

# ── Qwen2.5-7B-Instruct — Public Completions API ─────────────────────
# Qwen2.5-7B-Instruct model via a public completions endpoint (not chat).
# Uses prompt-style API (no messages array), so multi-turn is not
# supported. Good for testing raw completion-based apps.

resource "prisma-airs_red_team_target" "qwen_completions" {
  name            = "Test 123"
  target_type     = "APPLICATION"
  connection_type = "CUSTOM"

  connection_params = jsonencode({
    api_endpoint = "https://uoft-qwen2-5-7b-instruct-8828s.paas.ai.telus.com/v1/completions"
    request_headers = {
      "Content-Type"  = "application/json"
      "Authorization" = "Bearer ${var.qwen_api_key}"
    }
    request_json = {
      model      = "Qwen/Qwen2.5-7B-Instruct"
      prompt     = "{INPUT}"
      max_tokens = "300"
    }
    response_json = {
      choices = [
        { text = "{RESPONSE}" }
      ]
    }
    response_key = "text"
    multi_turn_config = {
      type           = "stateless"
      assistant_role = "assistant"
    }
  })
}

# ── Talkdesk Virtual Agent — Network Broker ───────────────────────────
# Talkdesk virtual agent behind a custom proxy. Uses a non-standard
# request format with contact_name and subject fields. Single-turn only.

resource "prisma-airs_red_team_target" "talkdesk" {
  name            = "Talkdesk - Network Channel"
  target_type     = "APPLICATION"
  connection_type = "CUSTOM"

  connection_params = jsonencode({
    api_endpoint = "https://redteam.cdot.io/api/v1/talkdesk/prompt"
    request_headers = {
      "Content-Type" = "application/json"
      "X-Api-Key"    = var.talkdesk_api_key
    }
    request_json = {
      prompt       = "{INPUT}"
      contact_name = "Test"
      subject      = "Account help"
    }
    response_json = {
      response = "{RESPONSE}"
    }
    response_key = "response"
  })
}
