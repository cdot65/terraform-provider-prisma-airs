# Installation

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

## Terraform Registry

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.1"
    }
  }
}
```

Then run:

```bash
terraform init
```

## Building from Source

```bash
git clone https://github.com/cdot65/prisma-airs-provider.git
cd prisma-airs-provider
make install
```

This builds the provider binary and installs it to `~/.terraform.d/plugins/` for local development.

## Verify Installation

Create a minimal configuration to verify the provider loads correctly:

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.1"
    }
  }
}

provider "prisma-airs" {}
```

```bash
terraform init
terraform providers
```
