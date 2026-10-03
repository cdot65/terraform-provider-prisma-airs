# Release Notes

## v0.10.0 — Gateway workspaces and Skill Scanning

### Skill Scanning

- Adopt published Go SDK v0.8.1, adding three Skill Scanning resources and nine data sources under AI Supply Chain Security.
- Manage tenant instance configuration, individual policy-rule state, and trusted fingerprints through native HCL.
- Keep rule identities stable, restore captured policy baselines on destroy, replace immutable trust overrides, and independently confirm deletion.
- Add explicit Skill Scanning bases under `supply_chain`; invalid settings affect only Skill Scanning usage.
- Use a true write-only authorization-code input with a nonsecret change version (Terraform 1.11+); redact upstream error bodies and retain identities for recovery.
- Read scans, findings, chains, and nullable statistics as native Terraform objects. Scan execution/uploads/CSV remain SDK/CLI operations.

### Gateway workspaces

- Coordinate dedicated IAM scope creation, workspace creation and slug binding. Destroy archives the workspace before deleting and confirming the owned scope.
- Add explicit managed/external scope modes, metadata-only workspace discovery, import/adoption, partial-failure checkpoints and read-only archive replacement planning.
- Expose native defaults and policy fields; clear removed managed settings and retain unconfigured settings. Support one usage policy per workspace, matching live service behavior.
- Add Gateway IAM endpoint configuration; keep role grants, membership and operational inference outside this release.

## v0.9.0 — AI Gateway management (2026-10-03)

- Add 15 Gateway resources and 13 safe metadata data sources, with organization/workspace ownership and explicit service/user key routes.
- Introduce optional `gateway` data/admin endpoint settings using shared OAuth defaults.
- Use native HCL routing/configuration objects; reject embedded routing credentials, preserve whole documents and field changes, and track version IDs separately from resource IDs.
- Manage individual integration workspace bindings without replacing other workspace access.
- Preserve sensitive desired inputs and one-time outputs; verify deletion and deployment archival. Workspaces, IAM, scans, inference, counter resets and automatic credential rotation remain external.
- Publish exact Registry schemas, workflow and complete examples from the product catalog.


## v0.8.0 — Product modules and configuration (2026-10-02)

- **Breaking:** identify Runtime Security and Supply Chain Security resources and data sources with product prefixes. Red Teaming names remain unchanged. Follow the [state migration guide](../guides/migration.md) before upgrading.
- **Breaking:** move endpoint overrides into `runtime`, `red_team`, and `supply_chain` blocks. Shared OAuth credentials and existing endpoint environment variables remain supported.
- Own registration, SDK clients, lifecycle code, and guide metadata within each product module. Generate navigation, reference coverage, and Registry subcategories from the same catalog.
- Show AI Gateway as not yet implemented; Gateway configuration and functionality are the next milestone.
- Preserve native HCL, profile revision diffs, import-only customer apps, sensitive state, and archive/deletion semantics.

## v0.7.0 — SDK upgrade, native HCL, and Docusaurus (2026-10-02)

- Rebuild documentation with Docusaurus and the shared AIRS Harness design, native HCL examples, exact provider schemas, migration guidance, and desktop/tablet/mobile browser and pixel checks.

- Pin published `prisma-airs-go` v0.6.1; correct typed missing/error handling, verified deletes, Model Security tombstones, and prompt-set archival.
- Track security profiles by logical name and highest numeric revision. Policy edits update the existing Terraform address with a new service UUID; rename leaves the previous name unmanaged. Destroy removes all history under the current name.
- **Breaking:** replace target JSON inputs with native HCL connection/authentication blocks. Preserve desired payloads and sensitive credentials; expose computed service discriminators. Databricks uses STREAMING; REST/STREAMING transports use the CUSTOM discriminator. Removing authentication requires replacement.
- **Breaking:** remove unsupported prompt-set `properties`; immutable key creation settings and Model Security source changes require replacement.
- **Breaking:** deployment-profile `profile_id`, `auth_code` and `details` are sensitive; outputs referencing them must set `sensitive = true`. Customer apps reject unsupported renames and explicit empty metadata. Custom topics reject empty descriptions, and clearing a prompt-set description plans replacement. Native provider connection changes plan replacement. DLP policy entries require nonempty `log_severity`, matching the live API.
- Verify API-key one-time secret preservation and import limitations, app/key cascading cleanup, and import-only customer-app behavior.
- Customer-app updates use the published SDK v0.6.1 deployment-code resolution fix. Provider and SDK version numbers are independent.

## v0.6.3 — Read-After-Write, Update Topic Resolution

- Fix: `Create` and `Update` now read back the full profile via `GetByID` — resolves "inconsistent result after apply" errors for `app_protection`, `database_security`, and `member.version` fields
- Fix: `resolveTopicRefs()` now runs on `Update` (was only on `Create`) — topics no longer become `null` after profile updates
- Update Truffles Agent example to match production config (`app_protection`, `database_security`, `member.version`)

## v0.6.2 — Auto-Resolve Topic Names

- **Feature:** Security profiles now auto-resolve `topic_name` to `topic_id` and `revision` via the Topics API before create/update — users no longer need to specify `topic_id` manually
- Error with actionable message if a referenced topic name doesn't exist

## v0.6.1 — SDK v0.4.1, Bool Serialization Fix

- Upgrade `prisma-airs-go` SDK to v0.4.1
- Fix: `bool` attributes set to `false` (e.g., `mask_data_in_storage = false`) are now correctly sent to the API instead of being silently dropped
- Fix: `slack_moderation` example uses `default_url_category` instead of `block_url_category` to match production profile
- Update example version constraint to `~> 0.6`

## v0.6.0 — SDK v0.4.0, New Security Profile Fields, .env Support

- Upgrade `prisma-airs-go` SDK to v0.4.0 (`aisec/management` + `aisec/scan` consolidated into `aisec/runtime`)
- **New schema fields** for `prisma-airs_security_profile`:
    - `app_protection.default_url_category` — list of default URL categories
    - `app_protection.url_detected_action` — action when URL detected (`"block"` or `""`)
    - `app_protection.malicious_code_protection` — nested block with `name` and `action`
    - `data_protection.database_security` — list of CRUD operation rules (`name` + `action`)
- Fix `mask_data_in_storage` state consistency (now Optional+Computed, always set in state)
- Remove hardcoded credentials from example files
- Add `scripts/terraform-env.sh` helper for `.env`-based credential loading
- Update docs with new fields, `.env` workflow, and authentication guide

## v0.5.0 — Customer App Import-Only, SDK v0.3.1

- **Breaking:** `prisma-airs_customer_app` no longer supports `terraform apply` for new apps — use `terraform import` instead. The AIRS API has no POST endpoint for customer apps; they are created externally.
- Upgrade `prisma-airs-go` SDK to v0.3.1 (removes `Create`, fixes List endpoint, adds new fields)
- Add computed attributes: `agent_app`, `ai_sec_profile_name`
- `tsg_id` is now computed-only (set by the API, not the user)
- Update docs and sandbox examples to reflect import-only workflow

## v0.4.1 — SDK v0.3.0

- Upgrade `prisma-airs-go` SDK to v0.3.0 (docs/examples only, no API changes)

## v0.4.0 — SDK v0.2.1, ForceDelete, Schema Validators, Doc Overhaul

- Upgrade `prisma-airs-go` SDK to v0.2.1
- Use `ForceDelete` for security profile and custom topic deletion (removes JSON parse workaround)
- Add `ToxicContentAction` compound values for toxic-content model protection (`high:block, moderate:allow`, etc.)
- Add schema validation via `stringvalidator.OneOf` for all protection names and actions
- Add `terraform-plugin-framework-validators` dependency
- Comprehensive documentation audit: fix all resource/data source docs to match Go schemas
- Fix API key docs: add missing required fields (`auth_code`, `rotation_time_interval`, `rotation_time_unit`)
- Fix model security rules docs: remove non-existent filter arguments, correct attribute name (`rules` not `items`)
- Fix deployment profiles docs: add missing `auth_code` attribute
- Fix red team custom prompt set docs: add missing `properties`, `status`, `active`, `archive` attributes
- Fix security profile docs: add `alert_url_category` to app_protection reference
- Update all version references to `~> 0.4`

## v0.3.2 — Fix Profile ID Revision Handling

- Fix: handle API revision model that changes `profile_id` on update
- Profile reads now reconcile server-assigned IDs after update operations

## v0.3.1 — State Consistency Fixes

- Fix: state consistency bugs in security profile resource
- Ensure Terraform state stays in sync with API after create/update

## v0.3.0 — SDK v0.2.0, GetByID/GetByName Lookups

- Refactor: use SDK v0.2.0 `GetByID`/`GetByName` for profile lookups
- Upgrade `prisma-airs-go` SDK to v0.2.0

## v0.2.0 — Native HCL Security Profile Schema

- **Breaking:** replace `policy = jsonencode(...)` with native HCL blocks (`ai_security_profile`, `model_protection`, `agent_protection`, `data_protection`)
- Security profiles now use typed nested blocks instead of opaque JSON strings
- Enables Terraform plan diffs, validation, and auto-complete for all profile fields

## v0.1.0 — Initial Release

### Resources (7)

| Resource | Domain |
|----------|--------|
| `prisma-airs_security_profile` | Management |
| `prisma-airs_custom_topic` | Management |
| `prisma-airs_api_key` | Management |
| `prisma-airs_customer_app` | Management |
| `prisma-airs_model_security_group` | Model Security |
| `prisma-airs_red_team_target` | Red Team |
| `prisma-airs_red_team_custom_prompt_set` | Red Team |

### Data Sources (3)

| Data Source | Domain |
|-------------|--------|
| `prisma-airs_dlp_profiles` | Management |
| `prisma-airs_deployment_profiles` | Management |
| `prisma-airs_model_security_rules` | Model Security |

### Infrastructure

- Terraform Plugin Framework (not SDKv2)
- OAuth2 `client_credentials` authentication for all domains
- Environment variable fallback for all credentials
- GoReleaser with multi-platform builds and GPG signing
- GitHub Actions CI/CD (lint, test, docs deploy, release)
- MkDocs Material documentation site
- E2E test suite with cleanup utility
