# Release Notes

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
