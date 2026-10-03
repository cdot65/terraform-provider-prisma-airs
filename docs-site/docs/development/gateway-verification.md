# Gateway implementation verification

The v0.9.0 candidate adds 15 Gateway resources and 13 metadata data sources using published Go SDK v0.6.1. Existing products retain their v0.8.0 schemas and Terraform names. The product catalog owns registrations, endpoint metadata, navigation and generated Registry pages.

## Validation evidence

- `make check`: formatting, vet, lint and race-enabled tests pass.
- Live Gateway tests execute with SCM OAuth on an entitled tenant and existing workspaces. No workspace/IAM provisioning or inference traffic is used.
- All 15 resource types pass creation, import, stable subsequent plan and verified destruction. Mutable fields update at stable IDs; workspace bindings have immutable endpoints.
- Nine core resource families recover from independent remote deletion or archival. Deployment destruction and refresh use archived status.
- Config plan checks prove unchanged routing fields and resource ID remain known while the revision ID becomes computed; config leaf changes produce updates rather than replacement.
- Both binding families preserve another workspace's access when their own binding is created and destroyed. Parent fixtures are independently confirmed absent after tests.
- All 13 discovery routes execute live. A separate read-only, pagination-aware inventory finds zero active disposable fixtures; retained archived deployment records are documented lifecycle outcomes.
- Unit tests cover native null/empty/false/zero values, external config additions, authorization errors, unknown inputs, apply-time credential guards, sensitive desired settings, and retaining creation identity/one-time secrets when GET fails.
- Documentation validation includes generated schema/Registry freshness, complete HCL examples, TypeScript, Docusaurus, browser navigation, and nine shared Harness pixel comparisons.

Private live logs, source manifests and read-only audit evidence are kept outside the repository. Skipped offline acceptance tests are not counted as live evidence. The historical all-product shell E2E harness was not used; Gateway live acceptance exercised Terraform CLI and independent SDK reads.

## Review gate

Claude Code review is pending. Scores and any resulting corrections will be recorded here after the independent review.

## Limits

Routing documents use native HCL and reject recognized embedded credential fields; never place any secret in those visible documents. Sensitive desired inputs cannot recover masked remote configuration drift or imported secrets. Removing such inputs does not guarantee remote erasure; set a supported explicit value or deliberately replace the resource.

Workspaces/IAM, runtime execution, infrastructure provisioning, connectivity probes, counter resets and automatic credential rotation remain external. Integration model selection, MCP capabilities/access and guardrail MCP mapping synchronization are separate management extensions requiring ownership design and tests. Other product API gaps remain follow-up work. See the [Gateway workflow](../guides/gateway-workflow.md) for the precise shipped surface and [migration](../guides/migration.md) when upgrading older product names.
