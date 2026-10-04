# Protected Skill Scanning onboarding

This complete configuration shows tenant registration using provider v0.10.0 and a write-only authorization code. Use it only with explicit onboarding ownership. The [public Supply Chain project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-supply-chain-security) keeps this lesson optional and provides safer catalog/trust exercises first.

## Before you start

Use Terraform 1.11+, shared management credentials, Skill Scanning entitlement, and permission to manage the tenant instance. Set both `PANW_SKILL_SCANNING_DATA_ENDPOINT` and `PANW_SKILL_SCANNING_MGMT_ENDPOINT` for your region. Obtain the complete authorized onboarding payload, including every deployment profile you intend to retain.

```hcl
# Setup: Write-only provider inputs require Terraform 1.11 or later.
terraform {
  required_version = ">= 1.11.0, < 2.0.0"
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

# Authentication: Management credentials and service bases come from the environment.
provider "prisma-airs" {}

# Registration: Supply the full desired native onboarding object from a secure source.
variable "registration" {
  type = object({
    tenant_id            = string
    support_account_id   = string
    created_by           = string
    registration_details = any
  })
  sensitive = true
}

# Authorization: The ephemeral variable prevents a code from being saved in plan or state.
variable "auth_code" {
  type      = string
  sensitive = true
  ephemeral = true
  default   = null

  validation {
    condition     = var.auth_code == null || var.auth_code_version != null
    error_message = "Set auth_code_version when supplying an authorization code."
  }
}

variable "auth_code_version" {
  type    = number
  default = null
}

# Instance: PUT owns the complete payload; retain deletion protection.
resource "prisma-airs_supply_chain_skill_scanning_instance" "tenant" {
  tenant_id            = var.registration.tenant_id
  support_account_id   = var.registration.support_account_id
  created_by           = var.registration.created_by
  registration_details = var.registration.registration_details
  auth_code            = var.auth_code
  auth_code_version    = var.auth_code_version

  lifecycle {
    prevent_destroy = true
  }
}
```

## Import before updating an existing instance

Load `TF_VAR_registration` through a secure variable source as a native object, and load `TF_VAR_auth_code` only when deliberately changing that code. Keep the same ephemeral value available for both plan and apply. Supply a nonsecret version when changing a code; increment it with an omitted code only when intentionally clearing it.

```bash
terraform init
terraform validate
terraform import prisma-airs_supply_chain_skill_scanning_instance.tenant 'YOUR_SKILL_TENANT_ID'
terraform plan -out=onboarding.tfplan
terraform apply onboarding.tfplan
```

Import only reads. The first apply after import sends a complete PUT; a partial GET response cannot reconstruct this payload. Omitted deployment profiles can be deactivated. The dedicated authorization code is absent from plan/state, while other registration metadata persists as sensitive state.

## End the exercise without deleting onboarding

The guard blocks `terraform destroy`. To release management to the external owner, back up state securely, remove this resource from configuration, remove only its Terraform binding, and inspect a plan:

```bash
umask 077
terraform state pull > onboarding-backup.tfstate
terraform state rm prisma-airs_supply_chain_skill_scanning_instance.tenant
terraform plan
```

State removal does not delete the tenant. Keep the backup private. Use the [instance lifecycle guide](../resources/skill-scanning-instance.md) for authorized provisioning/deletion, code-version behavior, and permission limits. No individual scan upload, execution, or deletion occurs in this example.
