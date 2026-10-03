---
page_title: "prisma-airs_supply_chain_skill_scanning_rule (Resource)"
subcategory: "AI Supply Chain Security"
---

# prisma-airs_supply_chain_skill_scanning_rule Resource

# Skill Scanning Rule

Manages the state of one catalog rule in the tenant's singleton Skill Scanning policy. Use the **catalog rule UUID**, not the effective rule-instance UUID. Terraform updates keep a stable identity and expose the changed `state` without replacing the object.

```hcl
resource "prisma-airs_supply_chain_skill_scanning_rule" "secrets" {
  rule_uuid = var.skill_rule_uuid
  state     = "BLOCKING"
}
```

States are `DISABLED`, `ALLOWING`, and `BLOCKING`. Create captures the current effective state in computed `original_state`, then updates only this rule. Update sends a one-rule map, leaving other rules unchanged. If the rule has no effective instance, its catalog default supplies the initial state. Destroy restores that default as an explicit effective setting; the API has no operation for returning to an absent instance.

Destroy restores `original_state` and confirms the effective setting. It does not delete the catalog rule or the tenant's policy, because those deletion operations do not exist. Manage each tenant/rule pair in only one Terraform resource. Do not manage the same pair in several states or workspaces.

```bash
terraform import prisma-airs_supply_chain_skill_scanning_rule.secrets '<catalog-rule-uuid>'
```

Import captures the effective state at import time as the restoration baseline; it cannot recover a previous configuration's baseline. Replacing the rule UUID restores the old rule before adopting the new rule.

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

See [Skill Scanning workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/skill-scanning-workflow/) and the exact schema in [Supply Chain Security](https://cdot65.github.io/terraform-provider-prisma-airs/reference/).

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Terraform identity. |
| `name` | `string` | computed | — | Catalog rule name. |
| `original_state` | `string` | computed | — | Effective state captured at adoption. Destroy restores this state. Import captures the current state. |
| `rule_instance_uuid` | `string` | computed | — | Server policy rule-instance UUID; may change when the service recreates it. |
| `rule_uuid` | `string` | required | — | Catalog rule UUID, not rule-instance UUID. Manage each tenant/rule pair once. |
| `state` | `string` | required | — | Desired rule state. |
| `tsg_id` | `string` | computed | — | Tenant Service Group ID. |
