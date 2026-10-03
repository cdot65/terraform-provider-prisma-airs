---
title: Skill Scanning policy and trusted skills
---

# Skill Scanning policy and trusted skills

This complete configuration manages one rule and a trusted fingerprint. Supply shared credentials through `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID`. Select a rule from the catalog and a fingerprint you have reviewed. Destroy restores the adopted rule state and removes trust.

```hcl
terraform {
  required_providers {
    prisma-airs = { source = "cdot65/prisma-airs" }
  }
}

variable "rule_uuid" { type = string }
variable "skill_fingerprint" { type = string }

provider "prisma-airs" {
  supply_chain {
    skill_scanning_data_endpoint = "https://api.apps.paloaltonetworks.com/aiag/data"
    skill_scanning_mgmt_endpoint = "https://api.apps.paloaltonetworks.com/aiag/mgmt"
  }
}

data "prisma-airs_supply_chain_skill_scanning_rules" "catalog" {}

resource "prisma-airs_supply_chain_skill_scanning_rule" "policy" {
  rule_uuid = var.rule_uuid
  state     = "BLOCKING"
}

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

[Policy lifecycle](../resources/skill-scanning-rule.md) and [trusted-skill lifecycle](../resources/skill-scanning-override.md) explain baseline restoration, replacement, import, and deletion verification. This example does not upload or execute a skill scan.
