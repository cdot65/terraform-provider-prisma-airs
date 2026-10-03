# prisma-airs_supply_chain_skill_scanning_overrides schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Reads Skill Scanning overrides without starting a scan or changing tenant settings.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `fingerprint` | `string` | optional | — | Exact lowercase SHA-256 fingerprint. |
| `id` | `string` | computed | — | Discovery identity. |
| `limit` | `number` | optional | — | Page size from 1 to 100; omission uses the server default. |
| `q` | `string` | optional | — | Broad name or trusting-identity search; at least three characters. |
| `result` | `dynamic` | computed | yes | Native Terraform object containing the SDK response, including nulls and nested lists. Access fields directly (for example result.rules); no JSON decoding is needed. Scans, overrides and findings return one API page. Rules/effective rules use offset traversal until empty (at most 100 pages of 100); concurrent service edits can affect this non-snapshot inventory. |
| `skill_name` | `string` | optional | — | Filter by skill name. |
| `skip` | `number` | optional | — | Nonnegative page offset. |
| `trusted_by` | `string` | optional | — | Filter by trusted by. |
