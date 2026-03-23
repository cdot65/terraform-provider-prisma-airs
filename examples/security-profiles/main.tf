# ---------------------------------------------------------------------------
# Security Profiles at Scale
# ---------------------------------------------------------------------------
# Manages multiple AI security profiles with varying levels of protection.
# Demonstrates how to use Terraform to maintain a fleet of security profiles
# across different applications and use cases.
#
# Prerequisites:
#   1. Copy tfvars:         cp terraform.tfvars.example terraform.tfvars
#   2. Set provider credentials via env vars or .env file:
#        cp ../../.env.example .env   # then fill in values
#        source .env                  # or use: ../../scripts/terraform-env.sh plan
# ---------------------------------------------------------------------------

terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.6"
    }
  }
}

provider "prisma-airs" {}
