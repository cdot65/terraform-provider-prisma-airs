# SDK upgrade verification — 2026-10-02

The provider v0.7.0 release candidate on `chore/sdk-v0.6.0` was verified using published `prisma-airs-go v0.6.1`, with `GOWORK=off` and no module replacement. The evidence below records pre-release verification; these changes require HCL/state refactoring for native targets and removed prompt-set properties. Phase 3 resource expansion and scanner-key regeneration remain backlog.

## Published dependency

The owner approved the upstream customer-app SDK fix, merged as [SDK PR #98](https://github.com/cdot65/prisma-airs-go/pull/98), merge `d3d54685bcc65b7b684f0efb28ee12fd9c7ebd9c`. [SDK v0.6.1](https://github.com/cdot65/prisma-airs-go/releases/tag/v0.6.1) and [release workflow 36999888164](https://github.com/cdot65/prisma-airs-go/actions/runs/36999888164) are published/successful. SDK checks and documentation checks passed, including nine browser navigation and nine visual parity checks. The provider's final app tests use the released module, not the preceding SDK worktree candidate.

## Validation

Use `GOTOOLCHAIN=go1.25.6`, `GOPRIVATE='github.com/cdot65/*'`, and `GOWORK=off`.

- `make check build`: formatting, vet, lint, race tests and binary build passed. Offline acceptance skips do not count toward live closure.
- `TF_ACC=1 go test -race -count=1 -v -timeout 15m ./internal/provider -run 'TestAcc(CustomTopic|RedTeamCustomPromptSet|ApiKey|CustomerApp|SecurityProfile)Resource_'`: all eleven tests passed on the runtime/red-team tenant, 2026-10-02 UTC. Covers omitted/cleared descriptions, empty examples, archives, nine key replacement plans, secret preservation/import, app import/update/delete and all four profile tests, including rich policy/defaults and one-field delta preservation.
- `TF_ACC=1 go test -race -count=1 -v -timeout 15m ./internal/provider -run 'TestAccRedTeamTargetResource_'`: all five tests and eight family subtests passed on the runtime/red-team tenant. Includes seven connection families, Databricks OAuth, auth defaults/transitions, native credential changes, connection/auth removals and borrowed-channel replacement.
- `TF_ACC=1 go test -race -count=1 -v -timeout 15m ./internal/provider -run 'TestAccModelSecurity(GroupResource_|_entitlement)'`: all three tests passed on the licensed Model Security tenant. Separate entitlement is required.
- Real Terraform CLI checks passed: payload leaf diffs remain visible, auth/inherited payload secrets stay masked, false/zero/null/empty values survive apply, and subsequent plans are stable. Deployment-profile `profile_id`, `auth_code` and `details` are sensitive; unmarked credential outputs are rejected.
- The original implementation documentation gate passed before the Docusaurus rebuild. All three example roots and the E2E root passed `terraform validate` using the built provider via `dev_overrides`; changed examples passed `terraform fmt -check`. The legacy shell E2E harness was not executed; acceptance tests supply the live lifecycle evidence.

Credentials came from existing AIRS CLI tenant configuration into subprocess environments, never repository files. Test target credentials were dummy values; no inference/stream requests ran. For independent reruns, supply valid creator metadata or `AIRS_ACC_CREATED_BY`, a deployment profile, and `AIRS_ACC_NETWORK_BROKER_CHANNEL_UUID` pointing to an existing readable fixture.

## Evidence integrity and cleanup

`go run ./e2e/audit --prefix=<run-prefix>` inventories Runtime/Red Team fixtures without mutations. Use `--domain=modelsecurity` for licensed group inventory, `--tombstones=<comma-separated-recorded-IDs>` for terminal group verification, and `--channel=<existing-UUID>` for a borrowed channel read. The live cleanup tool removed a dedicated topic and both revisions of a disposable profile, verifying each absence.

Private live evidence retains commands, timestamps, source-hash manifests and cleanup IDs outside the published documentation. Each final suite uses the published SDK with no replacement. Read-only inventories count zero active test objects; prompt sets remain archived and groups remain tombstoned where permanent removal is unavailable. Existing Network Broker channel and deployment-profile fixtures are read-only. Recorded terminal IDs and tenant-specific evidence are available in the project knowledge workspace, not this public site.

## Implementation provenance

Independent implementation review findings and scores are retained with the private source-hash manifests and live evidence in the project knowledge workspace. Those reviews predate the separate documentation-site rebuild. The final profile follow-up additionally proves identity-aware DLP metadata, unit cases for changed/reordered/unknown references, and a live DLP-reference switch without carrying old names/file settings.
