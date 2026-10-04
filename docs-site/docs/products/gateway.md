# AI Gateway

AI Gateway management is available in provider **v0.10.0**: 16 resources and 15 metadata data sources, native HCL routing documents, explicit workspace bindings, and verified deletion/archival, and coordinated workspace/IAM ownership.

Start with the [workspace guide](../resources/gateway-workspace.md) and [complete workspace example](../examples/gateway-workspaces.md).

Start with the [Gateway workflow](../guides/gateway-workflow.md), [complete example](../examples/gateway.md), or [generated product inventory](../reference/index.md). Shared OAuth credentials and an existing Gateway workspace are required. See [provider settings](../reference/provider-configuration.md) for the `gateway` endpoint block.

See the workflow for ownership, secret/import limits, and the management extensions that remain outside this release.

## Discover upstream providers

Provider 0.11.0 adds an [AI provider catalog data source](../data-sources/gateway-ai-providers.md), bringing Gateway coverage to **16 resources and 16 data sources**. The [OpenAI GPT and Claude Opus example](../examples/gateway.md) uses exact catalog slugs instead of manually supplied provider-family UUIDs. Install provider 0.11.0 or later from the Terraform Registry.
