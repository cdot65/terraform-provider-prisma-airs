---
title: Skill Scanning Trusted Skill Override
---

# Skill Scanning Trusted Skill Override

Trusts an exact skill fingerprint with the API's `ALLOW` decision. Create records the returned override UUID; reads search a complete, bounded inventory; destroy removes the override and independently confirms its absence.

```hcl
# Trust: Approve one exact fingerprint; edits replace this immutable override.
resource "prisma-airs_supply_chain_skill_scanning_override" "trusted" {
  skill_name  = "approved-skill"
  fingerprint = var.skill_fingerprint
  trusted_by  = "security@example.com"
  reason      = "Reviewed and approved"
}
```

The API has no override update method. Changes to the name, fingerprint, trusting identity, reason, or original scan UUID require replacement. Terraform normally removes the old trust before creating the new override; `create_before_destroy` is inappropriate for the same fingerprint. Removing trust restores ordinary policy evaluation.

```bash
terraform import prisma-airs_supply_chain_skill_scanning_override.trusted '<override-uuid>'
```

Read does not interpret authorization failures, duplicate pages, missing UUIDs, or exhausted pagination budgets as deletion. Credentials and tenant must match the state being managed. If a malformed create receipt checkpoints an identity from another TSG, updates and deletes stop. Verify that UUID independently, then use the correct tenant credentials; if it belongs elsewhere, remove the incorrect local binding with `terraform state rm` and import only the verified intended UUID. Removing state alone does not delete remote trust.

Skill Scanning belongs to **AI Supply Chain Security**. Configure both base URLs in `supply_chain`, using shared OAuth credentials. The Go SDK package name remains `aisec/agentguard`; Terraform names and guides use Skill Scanning.

```hcl
provider "prisma-airs" {
  supply_chain {
    skill_scanning_data_endpoint = "https://api.apps.paloaltonetworks.com/aiag/data"
    skill_scanning_mgmt_endpoint = "https://api.apps.paloaltonetworks.com/aiag/mgmt"
  }
}
```

Alternatively set `PANW_SKILL_SCANNING_DATA_ENDPOINT` and `PANW_SKILL_SCANNING_MGMT_ENDPOINT`. The established SDK variables `PANW_AGENT_GUARD_*_ENDPOINT` are fallback aliases. No endpoint is inferred from credentials.

See [Skill Scanning workflow](../guides/skill-scanning-workflow.md) and the exact schema in [Supply Chain Security](../reference/index.md).
