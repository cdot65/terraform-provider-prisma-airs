# Setup: Declare the provider required by this configuration.
terraform {
  required_version = ">= 1.11.0, < 2.0.0"
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {}
