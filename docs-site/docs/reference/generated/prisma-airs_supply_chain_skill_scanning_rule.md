# prisma-airs_supply_chain_skill_scanning_rule schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages one Skill Scanning policy rule. Create/update sets its state; destroy restores its captured original state because the API has no rule delete.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | computed | — | Terraform identity. |
| `name` | `string` | computed | — | Catalog rule name. |
| `original_state` | `string` | computed | — | Effective state captured at adoption. Destroy restores this state. Import captures the current state. |
| `rule_instance_uuid` | `string` | computed | — | Server policy rule-instance UUID; may change when the service recreates it. |
| `rule_uuid` | `string` | required | — | Catalog rule UUID, not rule-instance UUID. Manage each tenant/rule pair once. |
| `state` | `string` | required | — | Desired rule state. |
| `tsg_id` | `string` | computed | — | Tenant Service Group ID. |
