# prisma-airs_supply_chain_skill_scanning_scan schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Reads Skill Scanning scan without starting a scan or changing tenant settings.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `fingerprint` | `string` | optional | — | Lowercase SHA-256 fingerprint for scan lookup. |
| `id` | `string` | computed | — | Discovery identity. |
| `result` | `dynamic` | computed | yes | Native Terraform object containing the SDK response, including nulls and nested lists. Access fields directly (for example result.rules); no JSON decoding is needed. Scans, overrides and findings return one API page. Rules/effective rules use offset traversal until empty (at most 100 pages of 100); concurrent service edits can affect this non-snapshot inventory. |
| `scan_uuid` | `string` | optional | — | Scan UUID. Supply either scan_uuid or fingerprint. |
