# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------
# Sensitive values — populate via terraform.tfvars (see terraform.tfvars.example)
# ---------------------------------------------------------------------------

variable "litellm_api_key" {
  type        = string
  sensitive   = true
  description = "Bearer token for LiteLLM API"
}

variable "worf_api_key" {
  type        = string
  sensitive   = true
  description = "API key for worf (local qwen3) endpoint"
}

variable "bedrock_access_id" {
  type        = string
  sensitive   = true
  description = "AWS access key ID for Bedrock"
}

variable "bedrock_access_secret" {
  type        = string
  sensitive   = true
  description = "AWS secret access key for Bedrock"
}

variable "qwen_api_key" {
  type        = string
  sensitive   = true
  description = "Bearer token for Qwen2.5 endpoint"
}

variable "talkdesk_api_key" {
  type        = string
  sensitive   = true
  description = "API key for Talkdesk proxy endpoint"
}
