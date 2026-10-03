# prisma-airs_supply_chain_skill_scanning_override schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a trusted Skill Scanning fingerprint override. Edits replace the override; destroy removes trust.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `decision` | `string` | computed | — | Trusted-skill decision (ALLOW). |
| `fingerprint` | `string` | required | — | Exact skill fingerprint. |
| `id` | `string` | computed | — | Terraform identity. |
| `original_scan_uuid` | `string` | optional | — | Original scan UUID. |
| `reason` | `string` | optional | — | Reason for trusting the skill. |
| `skill_name` | `string` | required | — | Trusted skill name. Override changes require replacement; the API has no update operation. |
| `trusted_by` | `string` | required | — | Identity trusting the skill. |
| `tsg_id` | `string` | computed | — | Tenant Service Group ID. |
| `updated_at` | `string` | computed | — | Last modification timestamp. |
