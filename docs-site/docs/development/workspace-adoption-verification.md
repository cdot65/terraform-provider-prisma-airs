# SDK adoption and Gateway workspace verification

The v0.10.0 candidate combines Skill Scanning with coordinated Gateway workspace/IAM ownership and published Go SDK v0.8.1. No local module replacement is used. The Go SDK prerequisite was independently reviewed by actual Claude Code: helper 9.4/9.4, tests/live 9.2/9.5, documentation/release 9.3/9.3 (Standards/Spec), then published with passing Actions and a fresh module-proxy installation.

## Contracts and service behavior

Workspace/IAM contracts were recovered from the TypeScript SDK and historical traffic; these IAM routes are outside the supplied Gateway OpenAPI. Fresh disposable tests verify IAM scope create/get/bind/delete and post-delete GET, admin workspace create/get/update/archive, and child config access using the intended account. No access policies or memberships are created.

- Workspace UUID, unique slug and scope name are distinct. Renaming updates the label at the same UUID/slug.
- Admin detail decorates the returned label with its icon. The provider separates that decoration from the configured label; list rows use the plain label.
- The live service permits one usage policy per workspace. Credit updates preserve its UUID. The policy's server-only ID/status fields are excluded from configurable HCL, avoiding whole-object diffs.
- Rate collections replace completely: changing two entries to one removes the omitted entry.
- Defaults clear with config_id:null and metadata:{}; usage clears with null, rates with [], icon with an empty string. Empty usage arrays return HTTP400. Blank descriptions preserve the old value. Unreadable allow_config_override is excluded.
- Clearing removes workspace policy associations; hidden backing-record deletion is not asserted. Workspace archive rows remain by design.
- Live list envelopes omit has_more. Metadata is readable, but these envelopes do not prove complete inventory/absence. Ambiguous POST recovery therefore remains conservative and may require explicit UUID recovery/import.

## Terraform verification

Callback tests exercise real SDK HTTP boundaries: managed/external ownership, scope collisions and correlation-token recovery, sharing protection, failed binding resumption, archive/cleanup retry checkpoints, import/adoption, archived replacement planning, ambiguous workspace reconciliation, incomplete inventory refusal, typed native settings/drift, ignored-clear detection and metadata sanitization.

Real Terraform protocol tests exercise native settings, label/credit updates, stable UUID/slug, no-change planning, managed import, settings removal and independent mock cleanup. A missing external scope blocks ordinary plans without replacement; explicit destroy archives the existing workspace with no IAM writes. A private actual Terraform CLI failure/recovery test also retains the failed binding identity, resumes at the same UUID with changed configuration, produces an empty plan and independently confirms archive/scope cleanup. All product modules are regression tested against the published SDK. Live instance CRUD for Skill Scanning remains permission-blocked; see [its separate verification record](skill-scanning-verification.md). The v0.8.1 read-only scan/finding/chain/statistics test skipped because redtail-prod had no completed scan; those live assertions were not executed. Existing scans are read-only fixtures; no scan execution or uploads are added to Terraform.

Live Terraform managed lifecycle passed creation/binding, child config permissions, rename and credit updates with stable identities, import, empty plans, settings clearing and independent cleanup. External-mode create/rename/import/destroy preserved the IAM snapshot; the test owner then cleaned its separate disposable fixture scope. Remote archival planned destroy-before-create replacement, cleaned the old owned scope and recreated fresh workspace/scope-token identities; both old/new workspaces were confirmed inactive and the final scope absent. SDK-adoption live Skill Scanning trust create/import/replacement/no-op/delete and unchanged-policy discovery/restoration also passed.

The first actual provider review identified missing external-scope replacement safety, list-label decoration and actionable ambiguous-create recovery instructions. Regression tests first reproduced the destructive external replacement and literal-prefix loss, then passed after fixing them. Refresh retains the service status, records missing-scope workflow state, and permits replacement only for owned scopes. List and detail normalization are separate. Manual recovery now includes backup, verified identity/token/dedication, owner binding, state removal and managed/external import paths. Actual Claude Code second review on 2026-10-03 passed all fourteen tasks on both axes, with a minimum score of 9.2/10 and no required findings. This is an independent read-only review of the frozen candidate. Publication and fresh Registry/deployed-site verification follow the review gate; they are not claimed complete here.

## Actual Claude Code review scores

| Task | Standards | Spec |
| --- | --- | --- |
| Published SDK adoption and product regression | 9.3 | 9.3 |
| Skill Scanning instance adoption | 9.2 | 9.2 |
| Skill Scanning rule adoption | 9.2 | 9.2 |
| Skill Scanning override adoption | 9.3 | 9.4 |
| Skill Scanning discovery and state boundaries | 9.2 | 9.2 |
| Gateway native workspace schema and validation | 9.2 | 9.3 |
| Managed scope/workspace orchestration | 9.3 | 9.3 |
| External scope isolation | 9.3 | 9.3 |
| Workspace updates and settings clearing | 9.2 | 9.3 |
| Workspace archival and verified scope cleanup | 9.3 | 9.3 |
| Import, partial failures and recovery | 9.2 | 9.2 |
| Read-only refresh, replacement and metadata discovery | 9.2 | 9.2 |
| Terraform protocol, race and live verification | 9.3 | 9.3 |
| Registry/Docusaurus guides, exact schemas and release preparation | 9.2 | 9.2 |

Private receipts, red/green tests, source hashes and actual Claude reports are retained in `/tmp/prisma-airs-provider-sdk080/`. Tenant credentials and signed URLs are excluded from repository artifacts. No final provider review score is inferred from historical 9.0 reviews.
