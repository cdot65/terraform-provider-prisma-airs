# ---------------------------------------------------------------------------
# Prisma AIRS Provider — End-to-End Validation
# ---------------------------------------------------------------------------
# All credentials come from environment variables (source ../.env first):
#   PANW_AI_SEC_API_KEY, PANW_AI_SEC_PROFILE_NAME
#   PANW_MGMT_CLIENT_ID, PANW_MGMT_CLIENT_SECRET, PANW_MGMT_TSG_ID
# ---------------------------------------------------------------------------

terraform {
  required_providers {
    prisma-airs = {
      source = "cdot65/prisma-airs"
    }
  }
}

# Provider configured entirely via environment variables.
provider "prisma-airs" {}

# Unique suffix to avoid name collisions across runs.
resource "terraform_data" "run_id" {}

locals {
  prefix = "e2e-tf"
}
