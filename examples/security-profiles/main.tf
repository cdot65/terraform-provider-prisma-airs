# ---------------------------------------------------------------------------
# Security Profiles at Scale
# ---------------------------------------------------------------------------
# Manages multiple AI security profiles with varying levels of protection.
# Demonstrates how to use Terraform to maintain a fleet of security profiles
# across different applications and use cases.
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

provider "prisma-airs" {
  client_id     = "terraform-provider@1533764915.iam.panserviceaccount.com"
  client_secret = "686a9e15-f28d-44c3-a32e-75d419d03cdb"
  tsg_id        = "1533764915"
}
