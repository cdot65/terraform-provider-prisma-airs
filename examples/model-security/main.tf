# ---------------------------------------------------------------------------
# Model Security
# ---------------------------------------------------------------------------
# Manages model security groups for monitoring AI models from various
# sources. Demonstrates security group creation and reading model
# security rules.
#
# Prerequisites:
#   1. Copy tfvars:         cp terraform.tfvars.example terraform.tfvars
#   2. Set provider credentials (env vars or inline)
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
