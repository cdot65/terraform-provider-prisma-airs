# Released example verification

The provider 0.10.0 examples and documentation were independently reviewed using actual Claude Code CLI 2.1.288 on 2026-10-04. Its final verdict is **Ship**, with **9.1/10 overall**, **9.0/10 Standards**, and **9.4/10 Spec**.

The [complete review report](https://github.com/cdot65/prisma-airs-terraform-examples/blob/8ba428b3796c1b07b6a1fd5a8a5eb5d11fc0e7da/docs/reviews/provider-0.10.0.md) preserves the actual scores, four nonblocking P3 findings, review scope and verification boundaries. The reviewer independently ran the read-only documentation/hash checker. A subsequent Terraform 1.16.4 console check confirmed `nonsensitive(null)` returns null; the complete degraded-response path was not exercised by that check.

Reviewed source: public examples `40cea2ad2f149706dcd355d89cdb476bcfc99059` and provider documentation `5067e1dcb26d71bb389ec8cb7cb864b93b9e2589`. This page is a later evidence-only addition. No provider runtime behavior or schema changed.

The local checks passed five provider-checkout roots, 63 Registry pages, 54 exact schemas, 98 HCL snippets including eight complete configurations, TypeScript/production build, 11 browser tests and nine visual comparisons. The companion projects cover all 26 resources and 27 data sources, with 17 mock Terraform feature tests, 13 Python helper tests, exact source-hash/Markdown checks, and five fresh live lifecycle variants followed by independent cleanup.

Shared Skill onboarding/policy writes, external-secret retrieval, new hybrid installation, fresh MCP invocation, and external customer-app adoption remain unexercised. See the [precise live report](https://github.com/cdot65/prisma-airs-terraform-examples/blob/40cea2ad2f149706dcd355d89cdb476bcfc99059/docs/live-runs/provider-0.10.0.md) for the published receipts and boundaries.
