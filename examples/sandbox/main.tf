# ---------------------------------------------------------------------------
# Prisma AIRS Provider — Sandbox
# ---------------------------------------------------------------------------
# Interactive Terraform project for manually testing all resources and
# data sources. Use standard terraform commands:
#
#   terraform plan
#   terraform apply
#   terraform destroy
#
# Prerequisites:
#   1. Build the provider:  make build
#   2. Export credentials:
#        export PANW_MGMT_CLIENT_ID="..."
#        export PANW_MGMT_CLIENT_SECRET="..."
#        export PANW_MGMT_TSG_ID="..."
#      Or source the repo's .env file (see below).
# ---------------------------------------------------------------------------

terraform {
  required_providers {
    prisma-airs = {
      source = "cdot65/prisma-airs"
    }
  }
}

# Provider configured via environment variables.
# Override inline if you prefer:
#   provider "prisma-airs" {
#     client_id     = "..."
#     client_secret = "..."
#     tsg_id        = "..."
#   }
provider "prisma-airs" {}
