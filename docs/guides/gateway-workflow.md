---
page_title: "AI Gateway workflow"
---

# AI Gateway workflow

Provider v0.9.0 introduces 15 Gateway resources and 13 metadata data sources, covering CRUD for the twelve SDK families plus service/user key separation and two explicit workspace-binding resources. See the [derived product catalog](https://cdot65.github.io/terraform-provider-prisma-airs/reference/) for exact names and schemas.

## Configure access

Use the shared SCM OAuth credentials and TSG defaults. The optional `gateway` block accepts `data_endpoint` and `admin_endpoint`; otherwise `PANW_AI_GW_DATA_ENDPOINT`, `PANW_AI_GW_ADMIN_ENDPOINT` and SDK defaults apply. A Gateway-enabled tenant and existing workspace are prerequisites. Workspaces and IAM are externally managed.

## Establish dependencies

Create an organization integration, then its workspace-binding resource, then a workspace provider or MCP server. Declare `depends_on` on the binding. Bindings manage only their own integration/workspace pair; they never replace every workspace mapping or create default providers. Import existing enabled pairs before managing them.

See the [complete Gateway example](https://cdot65.github.io/terraform-provider-prisma-airs/examples/gateway/). Use `airs cli aigateway workspaces list` and `airs cli aigateway integrations providers` for discovery of externally managed workspaces and provider-family identifiers.

## Review routing changes

Write complete native HCL routing objects. Config updates retain the resource ID and change `version_id`; unchanged fields stay visible and known in the plan. The API replaces the whole document, so include settings you intend to retain. Remote document changes are refreshed as drift. Reference existing provider/integration or secret-reference identifiers rather than embedding plaintext credentials. JSON-string inputs and recognized credential fields are rejected.

## Preserve secrets

Guardrail parameter maps, integration/MCP configuration, manager auth and deployment inputs are sensitive desired settings; masked reads cannot detect arbitrary drift inside those values. Removing a sensitive input does not erase remote credentials. Supply an explicit new value or deliberately replace the object. API keys and deployment creation outputs are one-time material: preserve them in a secret store, protect state, and never expect import to recover them. Key or deployment auth rotation is not automatic.

## Destroy owned objects

Providers and servers depend on their workspace bindings and are removed first. Binding destruction disables only the owned pair. Deployment destruction archives the registration and verifies its status; discovery may still list archived records. Other resources verify remote absence. Externally archived/deleted resources leave managed state on refresh.

The provider does not perform inference, MCP tool invocation, scans, connectivity probes, connection termination, usage-counter resets, infrastructure/workspace/IAM provisioning, integration model selection, MCP capabilities/access synchronization, or guardrail MCP mapping synchronization. Those lifecycle/configuration extensions need their own ownership designs and tests. Other product API gaps remain separate follow-up work.
