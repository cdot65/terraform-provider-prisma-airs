# Gateway implementation verification

The v0.9.0 candidate adds 15 Gateway resources and 13 metadata data sources using published Go SDK v0.6.1. Existing products retain their v0.8.0 schemas and Terraform names. The product catalog owns registrations, endpoint metadata, navigation and generated Registry pages.

## Validation evidence

- `make check`: formatting, vet, lint and race-enabled tests pass.
- Live Gateway tests execute with SCM OAuth on an entitled tenant and existing workspaces. No workspace/IAM provisioning or inference traffic is used.
- The complete final-source live suite passes all 15 resource types: creation, import, stable subsequent plan and independently verified destruction. Its source manifest matches every current Go file and the published SDK dependency. Mutable fields update at stable IDs; workspace bindings have immutable endpoints.
- Nine core resource families recover from independent remote deletion or archival. Deployment destruction and refresh use archived status.
- The complete authored example passes apply and explicitly asserted empty subsequent plans, provider-note edits without routing revisions, routing leaf edits with new revisions, and verified destroy. Config plan checks prove unchanged routing fields and resource ID remain known while the revision ID becomes computed; config leaf changes produce updates rather than replacement.
- Both binding families preserve another workspace's access when their own binding is created and destroyed. Parent fixtures are independently confirmed absent after tests.
- All 13 discovery routes execute live and find freshly created owned fixture IDs. HCL page 1 maps to API page 0. A subsequent read-only inventory using corrected API page indices finds zero active disposable fixtures in all 13 families and retained archived deployment records. Earlier audits using incorrect page indices are superseded; eight recorded graph fixtures from failed runs were deleted and independently verified absent.
- Unit tests cover native null/empty/false/zero values, external config additions, authorization errors, unknown inputs, apply-time credential guards, sensitive desired settings, partial Dynamic/map/tuple/set plans with server-added fields, equivalent expiry timestamps, nonempty secret-reference mappings, explicit default-deployment requests, missing access-policy warnings, and retaining creation identity/one-time secrets when GET fails.
- Documentation validation includes generated schema/Registry freshness, complete HCL examples, TypeScript, Docusaurus, browser navigation, and nine shared Harness pixel comparisons.

Private live logs, source manifests and read-only audit evidence are kept outside the repository. Skipped offline acceptance tests are not counted as live evidence. The historical all-product shell E2E harness was not used; Gateway live acceptance exercised Terraform CLI and independent SDK reads.

## Independent review

Claude Code review led to corrections in planned-value reconciliation, stable metadata, credential detection, workspace isolation, scoped imports, pagination and evidence provenance. Review outcomes and task scores are recorded in [implementation PR #63](https://github.com/cdot65/terraform-provider-prisma-airs/pull/63).

## Limits

Routing documents use native HCL and reject recognized embedded credential fields; never place any secret in those visible documents. Sensitive desired inputs cannot recover masked remote configuration drift or imported secrets. Removing such inputs does not guarantee remote erasure; set a supported explicit value or deliberately replace the resource.

Workspaces/IAM, runtime execution, infrastructure provisioning, connectivity probes, counter resets and automatic credential rotation remain external. Inline provider rate-limit settings are excluded after repeatable upstream 503 responses on create; standalone rate policies are tested. Integration model selection, MCP capabilities/access and guardrail MCP mapping synchronization are separate management extensions requiring ownership design and tests. Other product API gaps remain follow-up work. See the [Gateway workflow](../guides/gateway-workflow.md) for the precise shipped surface and [migration](../guides/migration.md) when upgrading older product names.
