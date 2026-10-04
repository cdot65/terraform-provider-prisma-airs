# Model Security

Create a Hugging Face Model Security group and read the rule catalog. This configures an empty group; model onboarding and scanning are separate operations.

## Before you start

Load the [OAuth environment variables](../getting-started/authentication.md), install the [released provider](../getting-started/installation.md), and use a tenant with Model Security entitlement and management access. Choose an unused group name.

## Configure and apply

Save this complete configuration as `main.tf` in a separate directory:

```hcl
# Setup: Declare the provider required by this configuration.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {}

# Model group: Organize models; onboarding and scans are separate operations.
resource "prisma-airs_supply_chain_security_group" "models" {
  name        = "terraform-models"
  description = "Models managed through Terraform"
  source_type = "HUGGING_FACE"
}

# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_supply_chain_security_rules" "catalog" {}

# Outputs: Expose results for the next configuration or application step.
output "rule_names" {
  value = data.prisma-airs_supply_chain_security_rules.catalog.rules[*].name
}
```

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
```

## Update and clean up

Edit the group description to explore an update. Changing source_type instead plans replacement. Review and apply the change, then use `terraform plan -detailed-exitcode` to verify convergence.

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Destroy leaves the group tombstoned in the service; Terraform treats it as absent. Protect state and saved plans throughout the lifecycle.

For a complete working directory and recorded results, follow the [product project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-supply-chain-security).
