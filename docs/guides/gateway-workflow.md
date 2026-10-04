---
page_title: "AI Gateway workflow"
---

# AI Gateway workflow

Provider v0.10.0 includes 16 Gateway resources and 15 metadata data sources, covering CRUD for the twelve SDK families plus service/user key separation and two explicit workspace-binding resources. See the [derived product catalog](https://cdot65.github.io/terraform-provider-prisma-airs/reference/) for exact names and schemas.

## Configure access

Use the shared SCM OAuth credentials and TSG defaults. The optional `gateway` block accepts `data_endpoint`, `admin_endpoint` and `iam_endpoint`; otherwise `PANW_AI_GW_DATA_ENDPOINT`, `PANW_AI_GW_ADMIN_ENDPOINT`, `PANW_IAM_ENDPOINT` and SDK defaults apply. A Gateway-enabled tenant and appropriate existing role grants are prerequisites. Use the [managed workspace guide](https://cdot65.github.io/terraform-provider-prisma-airs/resources/gateway-workspace/) to coordinate a dedicated scope and workspace, or supply an existing externally managed workspace.

## Establish dependencies

Create an organization integration, then its workspace-binding resource, then a workspace provider or MCP server. Declare `depends_on` on the binding. Bindings manage only their own integration/workspace pair; they never replace every workspace mapping or create default providers. Import existing enabled pairs before managing them. If workspace-list metadata reports truncation and the pair is not visible, Terraform returns a verification error and retains state rather than declaring the binding absent.

See the [complete Gateway example](https://cdot65.github.io/terraform-provider-prisma-airs/examples/gateway/). Use `airs cli aigateway workspaces list` to locate your existing workspace. Resolve upstream provider-family IDs with the [AI provider catalog data source](https://cdot65.github.io/terraform-provider-prisma-airs/data-sources/gateway-ai-providers/), using `open-ai` for OpenAI and `anthropic` for Anthropic. Catalog discovery requires provider 0.11.0 or later.

## Review routing changes

Write complete native HCL routing objects. Config updates retain the resource ID and change `version_id`; unchanged fields stay visible and known in the plan. The API replaces the whole document, so include settings you intend to retain. Remote document changes are refreshed as drift. Reference existing provider/integration or secret-reference identifiers rather than embedding plaintext credentials. JSON-string inputs and recognized credential fields are rejected.

## Preserve secrets

Guardrail parameter maps, integration/MCP configuration, manager auth and deployment inputs are sensitive desired settings; masked reads cannot detect arbitrary drift inside those values. Removing a sensitive input does not erase remote credentials. Supply an explicit new value or deliberately replace the object. API keys and deployment creation outputs are one-time material: preserve them in a secret store, protect state, and never expect import to recover them. Key or deployment auth rotation is not automatic.

## Recover a failed follow-up read

A successful create saves the resource ID and available one-time outputs before its detail read. If that read fails, Terraform may mark the object tainted. Verify the object remotely first. If it is correct, run `terraform untaint <resource-address>` before applying again; otherwise a replacement can delete an API key or archive a deployment and issue new credentials. Fix noncanonical configuration values reported by reconciliation errors before retrying. The recovery checkpoint prioritizes identity and one-time outputs; a plan after untaint may re-send configured optional values that the create receipt did not echo. Review that plan before applying. A successful update followed by a failed read is already applied remotely; its computed fields reconcile on the next refresh.

## Discover metadata

Data sources return a page of safe metadata. Config and deployment routes expose no paging in the current SDK. An incomplete-list warning appears when their reported total exceeds the returned records; never use a page to prove absence. Other list sources accept one-based HCL pages, mapped to the API's zero-based index.

## Destroy owned objects

Providers and servers depend on their workspace bindings and are removed first. Binding destruction disables only the owned pair. Deployment destruction archives the registration and verifies its status; discovery may still list archived records. Other resources verify remote absence. Externally archived/deleted child objects leave managed state on refresh. Managed workspaces retain archived identities for owned-scope cleanup before replacement.

The provider does not perform inference, MCP tool invocation, scans, connectivity probes, connection termination, usage-counter resets, infrastructure provisioning, IAM access-policy grants and workspace membership, integration model selection, MCP capabilities/access synchronization, or guardrail MCP mapping synchronization. Those lifecycle/configuration extensions need their own ownership designs and tests. Other product API gaps remain separate follow-up work.
