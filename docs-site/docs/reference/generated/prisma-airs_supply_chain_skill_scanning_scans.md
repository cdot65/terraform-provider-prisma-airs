# prisma-airs_supply_chain_skill_scanning_scans schema

Exact data source attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Reads Skill Scanning scans without starting a scan or changing tenant settings.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `artifact_types` | `list(string)` | optional | — | Repeated artifact type filters. |
| `end_time` | `string` | optional | — | RFC3339 end time. |
| `fingerprint` | `string` | optional | — | Exact fingerprint filter. |
| `id` | `string` | computed | — | Discovery identity. |
| `limit` | `number` | optional | — | Page size from 1 to 100; omission uses the server default. |
| `result` | `dynamic` | computed | yes | Native Terraform object containing the SDK response, including nulls and nested lists. Access fields directly (for example result.rules); no JSON decoding is needed. Scans, overrides and findings return one API page. Rules/effective rules use offset traversal until empty (at most 100 pages of 100); concurrent service edits can affect this non-snapshot inventory. |
| `search_query` | `string` | optional | — | Search by scan name or UUID. |
| `skip` | `number` | optional | — | Nonnegative page offset. |
| `sort_order` | `string` | optional | — | Creation-time sort order. |
| `start_time` | `string` | optional | — | RFC3339 start time. |
| `statuses` | `list(string)` | optional | — | Repeated status filters. |
