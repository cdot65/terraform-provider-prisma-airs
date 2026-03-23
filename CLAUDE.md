# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Terraform provider for Palo Alto Networks Prisma AIRS. Built on the [prisma-airs-go](https://github.com/cdot65/prisma-airs-go) SDK. Covers all four service domains: AI Runtime Security (Scan), Management, Model Security, and AI Red Teaming.

## Commands

```bash
make fmt            # gofmt -s -w .
make vet            # go vet ./...
make lint           # golangci-lint run ./...
make test           # go test -race ./...
make test-coverage  # go test -race -coverprofile=coverage.out ./...
make testacc        # TF_ACC=1 acceptance tests (requires real credentials)
make build          # go build -o terraform-provider-prisma-airs
make check          # fmt + vet + lint + test (CI equivalent)
make install        # install provider locally for terraform development
make generate       # terraform-plugin-docs generate
make docs-serve     # mkdocs serve
```

Run a single test:

```bash
go test -v ./internal/provider/ -run TestSecurityProfileResource
go test -v ./... -run "TestName"
```

Note: `GOPRIVATE=github.com/cdot65/*` is set in the Makefile. For manual go commands, export it or prefix the command.

## Architecture

**Terraform Plugin Framework** (not SDKv2) — uses `terraform-plugin-framework` for typed schemas, plan modifiers, and validators.

```
main.go                         # provider server entry point
internal/provider/
  provider.go                   # provider schema, config, client init
  provider_test.go              # acceptance test helpers
  # Resources (CRUD):
  resource_security_profile.go  # management: security profiles
  resource_custom_topic.go      # management: custom topics
  resource_api_key.go           # management: API keys
  resource_customer_app.go      # management: customer apps
  resource_model_security_group.go
  resource_red_team_target.go
  resource_red_team_custom_prompt_set.go
  # Data Sources (read-only reference):
  datasource_dlp_profiles.go
  datasource_deployment_profiles.go
  datasource_model_security_rules.go
```

**SDK dependency:** `github.com/cdot65/prisma-airs-go` — private repo, requires `GOPRIVATE=github.com/cdot65/*`.

**Auth model:** Provider config → env var fallback. OAuth2 client_credentials for Management, Model Security, Red Team.

**Client initialization:** `provider.Configure()` resolves config, creates SDK clients. `ProviderData` struct is passed to all resources/data sources via `req.ProviderData`.

## Conventions

- Go 1.24+ minimum (go.mod targets 1.25)
- All resources implement `resource.Resource` + `resource.ResourceWithImportState`
- All data sources implement `datasource.DataSource`
- Test files: `*_test.go` alongside source, use `testAccProtoV6ProviderFactories`
- Acceptance tests gated by `TF_ACC=1` env var
- Resource naming: `prisma-airs_<resource>` (e.g., `prisma-airs_security_profile`)
- Schema field naming: snake_case matching Terraform conventions
- SDK model fields mapped to `types.String`, `types.Bool`, `types.Int64`, etc.

## CI/CD

- **ci.yml**: gofmt check, go vet, golangci-lint
- **test.yml**: `go test -race` matrix across Go versions
- **mkdocs-deploy.yml**: MkDocs Material → GitHub Pages on push to main
- **release.yml**: GoReleaser for Terraform provider binary distribution

## Docs

MkDocs Material site in `docs/`. Config in `mkdocs.yml`. Deployed to GitHub Pages at cdot65.github.io/prisma-airs-provider/.
