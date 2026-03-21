# Release Notes

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
