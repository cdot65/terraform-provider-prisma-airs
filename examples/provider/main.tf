terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.1"
    }
  }
}

# Configure via environment variables:
#   PANW_MGMT_CLIENT_ID
#   PANW_MGMT_CLIENT_SECRET
#   PANW_MGMT_TSG_ID
provider "prisma-airs" {}
