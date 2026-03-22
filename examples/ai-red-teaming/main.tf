# ---------------------------------------------------------------------------
# AI Red Teaming
# ---------------------------------------------------------------------------
# Manages red team targets and custom prompt sets for adversarial testing
# of AI applications. Demonstrates multiple connection types:
#   - CUSTOM targets via network broker (LiteLLM, Talkdesk)
#   - BEDROCK targets for AWS-hosted models (Claude)
#   - APPLICATION and MODEL target types
#
# Prerequisites:
#   1. Copy tfvars:         cp terraform.tfvars.example terraform.tfvars
#   2. Fill in API keys and credentials
#   3. Set provider credentials (env vars or inline)
# ---------------------------------------------------------------------------

terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.5"
    }
  }
}

provider "prisma-airs" {}
