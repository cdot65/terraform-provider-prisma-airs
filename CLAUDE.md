# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Terraform provider for Palo Alto Networks Prisma AIRS. Built on the [prisma-airs-go](https://github.com/cdot65/prisma-airs-go) SDK. Organized around AI Runtime Security, AI Red Teaming, AI Gateway, and AI Supply Chain Security. Gateway manages configuration objects; Supply Chain Security covers Model Security groups/rules and Skill Scanning registration, policy and trust.

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
make generate       # build provider + generate exact Docusaurus schema pages
make docs-serve     # Docusaurus development server
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
internal/provider/              # shared configuration, catalog composition, acceptance tests
internal/product/               # product definition and opaque client slots
internal/products/catalog.go    # product inventory, including Gateway
internal/products/runtime/      # Runtime Security lifecycle, data sources, SDK adapter
internal/products/gateway/     # Gateway management, bindings and metadata discovery
internal/products/redteam/      # Red Teaming lifecycle and native target inputs
internal/products/supplychain/  # Model Security and Skill Scanning, product-owned clients
internal/tfutil/                # shared SDK error/deletion handling
cmd/product-catalog/            # metadata for docs and schema consistency checks
```

**SDK dependency:** `github.com/cdot65/prisma-airs-go` — private repo, requires `GOPRIVATE=github.com/cdot65/*`.

**Auth model:** Provider config → env var fallback. OAuth2 client_credentials for Management, Model Security, Red Team.

**Client initialization:** `provider.Configure()` resolves shared defaults and nested product endpoints. Each product constructs its SDK client; opaque product-owned slots are passed via `req.ProviderData`.

## Conventions

- Go 1.25.6 minimum (see go.mod)
- All resources implement `resource.Resource` + `resource.ResourceWithImportState`
- All data sources implement `datasource.DataSource`
- Test files: `*_test.go` alongside source, use `testAccProtoV6ProviderFactories`
- Acceptance tests gated by `TF_ACC=1` env var
- Resource naming: `prisma-airs_<product>_<resource>` (e.g., `prisma-airs_runtime_security_profile`)
- Schema field naming: snake_case matching Terraform conventions
- SDK model fields mapped to `types.String`, `types.Bool`, `types.Int64`, etc.

## CI/CD

- **ci.yml**: gofmt check, go vet, golangci-lint
- **test.yml**: `go test -race` matrix across Go versions
- **deploy-docs.yml**: Docusaurus checks → GitHub Pages on push to main
- **release.yml**: GoReleaser for Terraform provider binary distribution

## Docs

Docusaurus site in `docs-site/`, authored guides in `docs-site/docs/`. For documentation changes, read `docs-site/docs/development/documentation.md` for schema generation and pinned Harness design checks. Complete `make docs-check` before declaring documentation changes verified. Existing public routes remain under cdot65.github.io/terraform-provider-prisma-airs/.

For Skill Scanning lifecycle changes, read `specs/skill-scanning-scope.md` for API ownership and `docs-site/docs/resources/skill-scanning-instance.md` for complete registration PUT semantics and write-only authorization codes.
