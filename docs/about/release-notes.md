# Release Notes

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
