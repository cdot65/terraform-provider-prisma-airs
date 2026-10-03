# Installation

## Requirements

Use Terraform 1.0 or later. Building the provider from source requires Go 1.25.6 or later; developing the documentation site requires Node 24+.

## Install from the Terraform Registry

Provider v0.9.0 uses Go SDK v0.6.1 and uses product-prefixed Terraform types and nested product endpoint blocks. Review [migration](../guides/migration.md) before upgrading from v0.7.0 or earlier.

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.9.0"
    }
  }
}

provider "prisma-airs" {}
```

```bash
terraform init
terraform providers
```

For an existing working directory, remove any development override and run `terraform init -upgrade`. Commit the resulting `.terraform.lock.hcl` with your configuration.

## Use the updated provider from source

Use a development override when testing unreleased changes:

```bash
git clone https://github.com/cdot65/terraform-provider-prisma-airs.git
cd terraform-provider-prisma-airs
make build
```

Create a Terraform CLI configuration file with the **absolute** path to the directory containing `terraform-provider-prisma-airs`:

```hcl
provider_installation {
  dev_overrides {
    "cdot65/prisma-airs" = "/absolute/path/to/terraform-provider-prisma-airs"
  }
  direct {}
}
```

Select it in your shell:

```bash
export TF_CLI_CONFIG_FILE=/absolute/path/to/dev.tfrc
```

For provider-only configurations, run `terraform validate`, `terraform plan`, and `terraform apply` directly. `dev_overrides` bypasses normal provider installation and registry version selection. Configurations with other providers or modules may still need `terraform init` to install those dependencies.

Provider and SDK release version numbers are independent.

## Verify the selected binary

```bash
terraform providers schema -json
terraform validate
```

The `prisma-airs_red_team_target` schema includes `openai`, `rest`, `streaming`, and the other [connection blocks](../resources/red-team-target.md#connection-blocks). It has no `connection_params` input.
