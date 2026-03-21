# Release Notes

## v0.1.0 — Initial Release

### Project Scaffolding
- Provider skeleton with Terraform Plugin Framework
- Provider schema with full auth configuration (API Key + OAuth2)
- Environment variable fallback for all credentials
- GoReleaser configuration for binary distribution

### CI/CD
- GitHub Actions: CI (lint, format, vet), Tests (Go 1.22–1.24 matrix)
- MkDocs Material documentation site with GitHub Pages deployment
- Release workflow with GoReleaser and GPG signing

### Documentation
- Getting started guides (installation, configuration, quick start)
- Resource documentation for all 9 resources
- Data source documentation for all 10 data sources
- Authentication guide
- Workflow guides (security profiles, model security, red team)
- Provider configuration reference
- Environment variable reference

### Planned Resources
- `prisma-airs_security_profile` — Management API
- `prisma-airs_custom_topic` — Management API
- `prisma-airs_api_key` — Management API
- `prisma-airs_customer_app` — Management API
- `prisma-airs_model_security_group` — Model Security API
- `prisma-airs_model_scan` — Model Security API
- `prisma-airs_red_team_target` — Red Team API
- `prisma-airs_red_team_scan` — Red Team API
- `prisma-airs_red_team_custom_prompt_set` — Red Team API

### Planned Data Sources
- `prisma-airs_content_scan` — Scan API
- `prisma-airs_dlp_profiles` — Management API
- `prisma-airs_deployment_profiles` — Management API
- `prisma-airs_scan_logs` — Management API
- `prisma-airs_model_security_rules` — Model Security API
- `prisma-airs_model_scan_evaluations` — Model Security API
- `prisma-airs_model_scan_violations` — Model Security API
- `prisma-airs_red_team_reports` — Red Team API
- `prisma-airs_red_team_categories` — Red Team API
- `prisma-airs_red_team_quota` — Red Team API
