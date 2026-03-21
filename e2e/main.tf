# ---------------------------------------------------------------------------
# Prisma AIRS Provider — End-to-End Validation
# ---------------------------------------------------------------------------
# All credentials come from environment variables (source ../.env first):
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
variable "run_id" {
  description = "Short unique ID appended to resource names to prevent collisions."
  type        = string
  default     = ""
}

locals {
  prefix = var.run_id != "" ? "e2e-tf-${var.run_id}" : "e2e-tf"
}
