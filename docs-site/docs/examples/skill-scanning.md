---
title: Skill Scanning policy and trusted skills
---

# Skill Scanning policy and trusted skills

Manage one tenant policy rule and one reviewed skill fingerprint with provider v0.10.0. This focused exercise changes shared policy: coordinate with the tenant owner before applying. For a first exercise that leaves policy unchanged, start with the [public Supply Chain project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-supply-chain-security) and its synthetic fingerprint lesson.

## Before you start

Load the [shared management credentials](../getting-started/authentication.md) and use a tenant entitled to Skill Scanning. Configure both service bases for your region; the configuration below shows the documented public bases. Scan execution and uploads remain CLI/SDK operations.

Set `rule_uuid` to a **catalog** rule UUID, distinct from its effective rule-instance UUID. Set `skill_fingerprint` to a reviewed lowercase SHA-256 fingerprint. Use a secure variable source and manage each tenant/rule pair in only one Terraform state.

## Configure and apply

Save this complete configuration as `main.tf`:

```hcl
# Setup: Install the released provider with shared write-only schema support.
terraform {
  required_version = ">= 1.11.0, < 2.0.0"
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

# Inputs: Select one catalog rule and one reviewed fingerprint.
variable "rule_uuid" {
  type = string
}
variable "skill_fingerprint" {
  type = string
}

# Authentication: Load credentials from the environment and choose service bases.
provider "prisma-airs" {
  supply_chain {
    skill_scanning_data_endpoint = "https://api.apps.paloaltonetworks.com/aiag/data"
    skill_scanning_mgmt_endpoint = "https://api.apps.paloaltonetworks.com/aiag/mgmt"
  }
}

# Discovery: Read existing results without starting scans or changing policy.
data "prisma-airs_supply_chain_skill_scanning_rules" "catalog" {}

# Policy: Adopt one shared rule; destroy restores its captured baseline.
resource "prisma-airs_supply_chain_skill_scanning_rule" "policy" {
  rule_uuid = var.rule_uuid
  state     = "BLOCKING"
}

# Trust: Approve one exact fingerprint; edits replace this immutable override.
resource "prisma-airs_supply_chain_skill_scanning_override" "trusted" {
  skill_name  = "reviewed-skill"
  fingerprint = var.skill_fingerprint
  trusted_by  = "security@example.com"
  reason      = "Approved following review"
}

output "catalog_rule_names" {
  value = [for rule in data.prisma-airs_supply_chain_skill_scanning_rules.catalog.result.rules : rule.name]
}
```

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
```

The rule captures its prior effective state and updates only that rule. The override trusts an exact fingerprint; the example does not prove that its contents are safe or launch analysis.

## Update and clean up

Change the desired rule state to practice an in-place policy update. Changes to the immutable trust fields, including reason, replace the override. Review a saved plan before applying either change.

```bash
terraform plan -out=update.tfplan
terraform apply update.tfplan
terraform plan -detailed-exitcode
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Destroy restores the captured rule baseline and removes trust. If no effective rule existed, restoration writes its catalog default explicitly; there is no API operation to restore an absent row. Import captures the state at import time, rather than recovering an earlier baseline. Read the [policy lifecycle](../resources/skill-scanning-rule.md) and [trusted-skill lifecycle](../resources/skill-scanning-override.md).

For existing scans, vulnerabilities, attack chains, nullable statistics, and protected tenant onboarding, follow the [complete product project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-supply-chain-security) or [onboarding lesson](skill-onboarding.md). Findings and registration metadata remain sensitive in state.
