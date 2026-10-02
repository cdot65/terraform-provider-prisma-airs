# Product refactor verification

The v0.8.0 candidate retains seven resources and three data sources while moving their implementation, SDK clients, registrations, and guide metadata into product modules. AI Gateway remains explicitly unimplemented.

## Offline checks

- Go formatting, vet, golangci-lint, and race tests passed with Go 1.25.6.
- All seven resource schemas and three data-source schemas exactly matched installed v0.7.0 after the intentional type renames. Resource fields, sensitivity, nested blocks, and schema versions did not change.
- Fifteen moved implementation files retained their lifecycle function bodies apart from public names and shared error qualifiers.
- Composed-provider tests passed for schema ownership, shared credentials, explicit nested overrides, omitted/partial/empty-block environment fallback, endpoint routing, client type isolation, and product-specific missing-credential diagnostics, and nil/invalid provider-data callbacks. Configure performed no network requests.
- Documentation checks validated 38 HCL examples, 18 Registry pages, and 11 schema pages. Nine browser tests and nine desktop/tablet/mobile pixel comparisons passed against the pinned Harness reference.

## Live acceptance and cleanup

All 22 unique acceptance tests have passing evidence using published Go SDK v0.6.1, `GOWORK=off`, and no module replacement. Runtime Security and Red Teaming used an authorized management tenant; Supply Chain Security used a separate licensed tenant.

Coverage includes keys and one-time secret preservation, import-only customer apps and remote deletion, custom topics, native target families/authentication/replacement, prompt-set archival, profile policy deltas/revisions/rename/history deletion, deployment/DLP catalogs, Model Security group lifecycle/replacement/tombstones, rules, and entitlement.

The broad Runtime/Red Team command initially failed because custom target creation returned HTTP 503. Its other tests passed; the focused custom-family retry passed without a code change. This is aggregate passing evidence, not a claim that the initial broad command exited successfully.

Independent read-only audits found zero active disposable fixtures across all tested domains, confirmed archived prompt sets; acceptance cleanup checks confirmed group tombstones, and verified the borrowed Network Broker channel remained present. Archive-only service records remain intentionally. No nonremovable channel was created.

## State migration rehearsal

A disposable v0.7.0 configuration with a security profile, topic, and DLP data source was installed directly from the Registry. The v0.8.0 side used the built candidate with a development override because it was not yet published.

Removing only one renamed resource reproduced a missing-schema error on plan while another old type remained in state. Removing all renamed managed/data-source addresses first, then importing the profile by name and topic by ID, preserved both remote identities and produced a no-change plan (exit 0). The fixtures were destroyed with the candidate and a read-only audit confirmed zero remaining active matches.

## Independent review and release gates

Claude Code review covers product modules/catalog, configuration/contract preservation, and documentation/migration/CI. The independent prerelease review passed on 2026-10-02 after correcting and rehearsing the migration sequence.

| Implementation task | Claude Code score |
| --- | --- |
| Product modules and catalog | 9/10 |
| Configuration and contract preservation | 9/10 |
| Documentation, migration, and CI | 9/10 |

All material findings were resolved. The review was read-only and evaluated source plus execution evidence. Publication requires at least 9/10 for each task, resolved material findings, final checks, CI, signed release assets, and Registry installation verification.

## Limits

This milestone does not implement Gateway, scan execution, scanner-key regeneration, IAM, or workspace provisioning. The shell E2E harness was not used; live acceptance and read-only audit provided the lifecycle evidence. Renamed Terraform types require explicit state migration; imports cannot recover one-time or masked secrets. See [migration](../guides/migration.md).
