---
title: Skill Scanning verification
---

# Skill Scanning verification

The implementation uses published Go SDK v0.7.0. Terraform-facing names are Skill Scanning under AI Supply Chain Security; SDK wire contracts and package identifiers retain their upstream names.

## Verification scope

Three resources and nine data sources are added. Mock Terraform lifecycle tests exercise instance CRUD/import, rule updates with stable identity and baseline restoration, override replacement/import/delete, native response traversal, and selector validation. Callback tests exercise failed-read identity checkpoints, permission failures, tenant ownership, repeated/clamped paging, deletion receipts, confirmed absence, unknown inputs, and precise null/false/number semantics.

Live read-only discovery on the configured live tenant returned seven catalog rules, seven effective rule instances, and no preexisting trusted-skill overrides. Instance GET returned HTTP 403; no live tenant provisioning, update, or deletion is claimed. The live tests use a synthetic disposable fingerprint and adopt a rule without changing its existing effective state. Policy state-change deltas and restoration after a changed value are covered by mocks rather than altering production policy.

Individual scans have no delete operation. The provider reads existing scans; it does not reserve signed URLs, upload archives, start/poll analysis, or export CSV. A separate existing SDK tenant supplies completed scans as read-only fixtures; the lifecycle-test tenant has no completed scans. Attack-chain detail remains mock-only when no live chain exists.

## Recorded checks

`make check` passes formatting, vet, lint, and all-package race tests. All twenty-two existing resource schemas and sixteen existing data-source schemas are unchanged; the candidate registers twenty-five resources and twenty-five data sources. Exact generation produces fifty-one Docusaurus schema pages and sixty Registry pages. A regression test and browser navigation verify separate instance resource/data-source schemas.

`make docs-check` passes Markdown checks, complete HCL validation, exact schema freshness, TypeScript, production build, ten navigation/browser checks, and all nine Harness pixel comparisons.

Live trusted-skill validation passed create/import/replacement/empty-plan/destroy with independently confirmed absence of both generated UUIDs. Live catalog/effective policy discovery, unchanged-policy adoption/import/empty-plan/restoration passed on the configured live tenant. Read-only scan lookup/detail/list, findings, attack chains and scan/rule statistics passed against the existing SDK tenant. No new scans or tenant instances were created.

Terraform tests also exercise normalized/omitted GET metadata, failed updates followed by a nonempty replan, the first full PUT after import, and server-created rule-instance UUID changes. Native registration tests preserve large numbers in entitlements, deployment profiles, extra metadata and extensions. Token-routing tests verify shared OAuth options and preserve existing Model Security fallback. Instance updates retain prior state after failed writes and resend complete desired registration details. Authorization codes are configuration-only inputs; other sensitive registration values require a protected Terraform state backend. Live instance CRUD remains unverified.
